package repo

import (
	"math/rand"
	"sync"
	"time"
)

type BoobaRepo struct {
	cache []int
	mutex sync.RWMutex
}

func (r *BoobaRepo) InitCache(content []int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.cache = content
}

func (r *BoobaRepo) GetBooba() (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	rand.Seed(time.Now().Unix())
	return r.cache[rand.Intn(len(r.cache))], nil
}
