package repo

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"slices"
	"strconv"
	"time"
)

//go:embed bayan.lua
var bayan string

const (
	clientsAllKey      = "clients:all"
	clientsPrivateKey  = "clients:private"
	statsLaunchesKey   = "stats:launches"
	statsUsersKey      = "stats:users:perDay"
	statsUsersTodayKey = "stats:users:today"
	lastLaunchKey      = "last:launch"
	dateFormat         = "06.02.01"
)

type BoobaRepo struct {
	rdb        *redis.Client
	ctx        context.Context
	launchName string
	bayanSHA   string
}

type ReqData struct {
	Date     string `json:"date,omitempty"`
	Requests int    `json:"requests"`
	Users    int    `json:"users"`
}

func (r *BoobaRepo) redisInit(content []int) {
	today := time.Now().Format(dateFormat)
	r.launchName = time.Now().Format(time.DateTime)

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

	lastLaunch, err := r.rdb.Get(r.ctx, lastLaunchKey).Result()
	if err != nil {
		lastLaunch = today
	}

	hornyUsers, err := r.rdb.SCard(r.ctx, statsUsersTodayKey).Result()

	pipe := r.rdb.TxPipeline()

	pipe.LPush(r.ctx, statsLaunchesKey, r.launchName)

	anyContent := make([]any, 0, len(content))
	for _, num := range content {
		anyContent = append(anyContent, num)
	}
	pipe.SAdd(r.ctx, "booba:new", anyContent...)
	pipe.Rename(r.ctx, "booba:new", "booba:active")
	if lastLaunch != today {
		pipe.HIncrBy(r.ctx, statsUsersKey, lastLaunch, hornyUsers)
		pipe.Del(r.ctx, statsUsersTodayKey)
	}
	pipe.Set(r.ctx, lastLaunchKey, today, 0)
	_, err = pipe.Exec(r.ctx)
	if err != nil {
		log.Fatal("error adding data to redis:", err)
	}
}

func (r *BoobaRepo) InitCache(content []int) {
	r.redisInit(content)
}

// GetBooba pops random post id from set
func (r *BoobaRepo) GetBooba() (int, error) {
	res, err := r.rdb.SPop(r.ctx, "booba:active").Result()
	if err != nil {
		return 0, err
	}
	intRes, err := strconv.Atoi(res)
	return intRes, err
}

func (r *BoobaRepo) IncViews() {
	err := r.rdb.HIncrBy(r.ctx, "stats:views", r.launchName, int64(1)).Err()
	if err != nil {
		log.Fatal("error incrementing views:", err)
	}
}

// AddBayan adds boobaID to list and set associated with chatID, while keeping both trimmed to maxLen size,
// uses ./bayan.lua script
func (r *BoobaRepo) AddBayan(chatID int, boobaID int, maxLen int) {
	strID := strconv.Itoa(chatID)
	set, list := "bayan:set:"+strID, "bayan:list:"+strID
	usersToday := statsUsersTodayKey
	//scripts are volatile in Redis, so error may be just "I forgot the script"
	err := r.rdb.EvalSha(r.ctx, r.bayanSHA, []string{set, list, usersToday}, boobaID, maxLen, strID).Err()
	if err != nil {
		r.bayanSHA, err = r.rdb.ScriptLoad(r.ctx, bayan).Result()
		if err != nil {
			log.Fatal("error loading script:", err)
		}

		err = r.rdb.EvalSha(r.ctx, r.bayanSHA, []string{set, list, usersToday}, boobaID, maxLen, strID).Err()
		if err != nil {
			log.Fatal("error executing script:", err)
		}
	}

}

// CheckBayan checks if post № boobaID was already posted to chatID
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
	pipe.SAdd(r.ctx, clientsAllKey, chatID)
	if chatID < 0 {
		pipe.SAdd(r.ctx, clientsPrivateKey, chatID)
	}
	_, err := pipe.Exec(r.ctx)
	if err != nil {
		log.Fatal("error adding chat №", chatID, " to redis:", err)
	}
}

func (r *BoobaRepo) GetStats() (int, int) {
	var chats, priv int64
	var err error

	chats, err = r.rdb.SCard(r.ctx, clientsAllKey).Result()
	if err != nil {
		log.Fatal("error getting size of user set:", err)
	}
	priv, err = r.rdb.SCard(r.ctx, clientsPrivateKey).Result()
	if err != nil {
		log.Fatal("error getting size of private set:", err)
	}

	return int(chats), int(priv)
}

func (r *BoobaRepo) GetRequests() []ReqData {
	views := map[string][2]int{}
	today := time.Now()
	t := today.AddDate(0, 0, -1)
	var dates []string
	start := t.AddDate(0, 0, -8)
	for ; t.After(start); t = t.AddDate(0, 0, -1) {
		dates = append(dates, t.Format(dateFormat))
	}
	slices.Reverse(dates)
	for _, date := range dates {
		views[date] = [2]int{0, 0}
	}

	keys, err := r.rdb.LRange(r.ctx, statsLaunchesKey, 0, 100).Result()
	if err != nil {
		log.Fatal("error getting keys for launches:", err)
	}

	var value string
	var nValue int
	var cur time.Time
	var ok bool
	var tmp [2]int

	for _, key := range keys {
		value, err = r.rdb.HGet(r.ctx, "stats:views", key).Result()
		if err != nil {
			value = "0"
		}
		nValue, err = strconv.Atoi(value)
		if err != nil {
			log.Fatal("error parsing value:", err)
		}

		cur, err = time.Parse(time.DateTime, key)
		if err != nil {
			log.Fatal("error parsing date:", err)
		}
		if _, ok = views[cur.Format(dateFormat)]; ok {
			tmp = views[cur.Format(dateFormat)]
			tmp[0] += nValue
			views[cur.Format(dateFormat)] = tmp
		}
	}

	for k := range views {
		value, err = r.rdb.HGet(r.ctx, statsUsersKey, k).Result()
		if err != nil {
			value = "0"
		}
		nValue, err = strconv.Atoi(value)
		if err != nil {
			log.Fatal("error parsing value:", err)
		}
		tmp = views[k]
		tmp[1] += nValue
		views[k] = tmp
	}

	var res []ReqData
	for _, v := range dates {
		res = append(res, ReqData{v[3:], views[v][0], views[v][1]})
	}
	return res
}
