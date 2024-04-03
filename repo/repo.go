package repo

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"strconv"
	"sync"
)

type BoobaRepo struct {
	cache []int
	mutex sync.RWMutex
	rdb   *redis.Client
	ctx   context.Context
}

func (r *BoobaRepo) redisInit(content []int) {
	r.ctx = context.Background()
	url := os.Getenv("REDIS_URL")
	opts, err := redis.ParseURL(url)
	if err != nil {
		log.Fatal("error getting data for redis connection:", err)
	}
	r.rdb = redis.NewClient(opts)

	anyContent := make([]any, 0, len(content))
	for _, num := range content {
		anyContent = append(anyContent, num)
	}
	err = r.rdb.SAdd(r.ctx, "booba_new", anyContent...).Err()
	if err != nil {
		log.Fatal("error adding booba to Redis:", err)
	}
	err = r.rdb.Rename(r.ctx, "booba_new", "booba_active").Err()
	if err != nil {
		log.Fatal("error renaming key:", err)
	}

	log.Print("Pushed into Redis!")
}

func (r *BoobaRepo) InitCache(content []int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.cache = content

	go r.redisInit(content)
}

func (r *BoobaRepo) GetBooba() (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	//	rand.Seed(time.Now().Unix())
	//	return r.cache[rand.Intn(len(r.cache))], nil
	res, err := r.rdb.SPop(r.ctx, "booba_active").Result()
	if err != nil {
		return 0, err
	}
	intRes, err := strconv.Atoi(res)
	return intRes, err
}
