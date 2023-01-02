package repo

import (
	"github.com/go-redis/redis/v8"
	"log"
	"math/rand"
	"os"
	"time"
)

type BoobaRepo struct {
	client *redis.Client
}

func (r BoobaRepo) InitCache(content []string) {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		log.Fatal("no Redis cache!")
	}

	opt, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}

	r.client = redis.NewClient(opt)

	for _, booba := range content {
		r.client.RPush(r.client.Context(), "booba", booba)
	}
}

func (r BoobaRepo) GetBooba() (string, error) {
	res, err := r.client.LRange(r.client.Context(), "booba", 0, -1).Result()
	if err != nil {
		return "", err
	}

	s := rand.NewSource(time.Now().Unix())
	g := rand.New(s)
	return res[g.Intn(len(res))], nil
}
