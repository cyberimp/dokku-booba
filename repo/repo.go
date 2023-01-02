package repo

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"math/rand"
	"os"
	"time"
)

type BoobaRepo struct {
	client *redis.Client
	ctx    context.Context
}

func (r BoobaRepo) InitCache(content []string) {
	r.ctx = context.Background()
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
		_, err := r.client.RPush(r.ctx, "booba", booba).Result()
		if err != nil {
			panic(err)
		}
	}
}

func (r BoobaRepo) GetBooba() (string, error) {
	res, err := r.client.LRange(r.ctx, "booba", 0, -1).Result()
	if err != nil {
		return "", err
	}

	rand.Seed(time.Now().Unix())
	return res[rand.Intn(len(res))], nil
}
