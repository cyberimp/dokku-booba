package repo

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strconv"
	"time"
)

//go:embed bayan.lua
var bayan string

type BoobaRepo struct {
	rdb        *redis.Client
	ctx        context.Context
	launchName string
	bayanSHA   string
}

func (r *BoobaRepo) redisInit(content []int) {
	r.launchName = time.Now().String()

	r.ctx = context.Background()
	url := os.Getenv("REDIS_URL")
	opts, err := redis.ParseURL(url)
	if err != nil {
		log.Fatal("error getting data for redis connection:", err)
	}
	r.rdb = redis.NewClient(opts)

	r.bayanSHA, err = r.rdb.ScriptLoad(r.ctx, bayan).Result()
	if err != nil {
		log.Fatal("error loading script:", err)
	}

	pipe := r.rdb.TxPipeline()

	pipe.LPush(r.ctx, "Launches", r.launchName)

	anyContent := make([]any, 0, len(content))
	for _, num := range content {
		anyContent = append(anyContent, num)
	}
	pipe.SAdd(r.ctx, "booba:new", anyContent...)
	pipe.Rename(r.ctx, "booba:new", "booba:active")
	_, err = pipe.Exec(r.ctx)
	if err != nil {
		log.Fatal("error adding data to redis:", err)
	}
}

func (r *BoobaRepo) InitCache(content []int) {
	r.redisInit(content)
}

func (r *BoobaRepo) GetBooba() (int, error) {
	res, err := r.rdb.SPop(r.ctx, "booba:active").Result()
	if err != nil {
		return 0, err
	}
	intRes, err := strconv.Atoi(res)
	return intRes, err
}

func (r *BoobaRepo) IncViews() {
	err := r.rdb.Incr(r.ctx, r.launchName).Err()
	if err != nil {
		log.Fatal("error incrementing views:", err)
	}
}

func (r *BoobaRepo) AddBayan(chatID int, boobaID int, maxLen int) {
	strID := strconv.Itoa(chatID)
	set, list := "bayan:set:"+strID, "bayan:list:"+strID
	err := r.rdb.EvalSha(r.ctx, r.bayanSHA, []string{set, list}, boobaID, maxLen).Err()
	if err != nil {
		r.bayanSHA, err = r.rdb.ScriptLoad(r.ctx, bayan).Result()
		if err != nil {
			log.Fatal("error loading script:", err)
		}
		err = r.rdb.EvalSha(r.ctx, r.bayanSHA, []string{set, list}, boobaID, maxLen).Err()
		if err != nil {
			log.Fatal("error executing script:", err)
		}
	}

}

func (r *BoobaRepo) CheckBayan(chatID int, boobaID int) bool {
	strID := strconv.Itoa(chatID)
	set := "bayan:set:" + strID
	res, err := r.rdb.SIsMember(r.ctx, set, boobaID).Result()
	if err != nil {
		log.Fatal("error checking for bayan:", err)
	}
	return res
}

func (r *BoobaRepo) AddChat(chatID int) {
	pipe := r.rdb.TxPipeline()
	pipe.SAdd(r.ctx, "clients:all", chatID)
	if chatID < 0 {
		pipe.SAdd(r.ctx, "clients:private", chatID)
	}
	_, err := pipe.Exec(r.ctx)
	if err != nil {
		log.Fatal("error adding chat №", chatID, " to redis:", err)
	}
}

func (r *BoobaRepo) GetStats() (int, int) {
	var chats, priv int64
	var err error

	chats, err = r.rdb.SCard(r.ctx, "clients:all").Result()
	if err != nil {
		log.Fatal("error getting size of user set:", err)
	}
	priv, err = r.rdb.SCard(r.ctx, "clients:private").Result()
	if err != nil {
		log.Fatal("error getting size of private set:", err)
	}

	return int(chats), int(priv)
}
