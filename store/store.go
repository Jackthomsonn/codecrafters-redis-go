package store

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackthomsonn/redis-go-impl/internal"
)

type RedisStore struct {
	Data map[string]RedisObject
	mu   sync.Mutex
}

type RedisObject struct {
	Value  interface{}
	Expire uint64
}

func NewRedisStore() *RedisStore {
	return &RedisStore{Data: make(map[string]RedisObject)}
}

func (rs *RedisStore) Set(res internal.ParsedResponse) {
	rs.mu.Lock()
	rs.Data[res.Key] = RedisObject{Value: res.Value, Expire: res.Mili}
	rs.mu.Unlock()
	go rs.handleRemovalOfExpiredData(res.Mili, res.Key)
}

func (rs *RedisStore) Get(key string) (interface{}, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	data, exists := rs.Data[key]
	if !exists {
		return nil, errors.New("key not found")
	}
	return data.Value, nil
}

func (rs *RedisStore) handleRemovalOfExpiredData(mili uint64, key string) {
	if mili == 0 {
		return
	}
	time.Sleep(time.Duration(mili) * time.Millisecond)
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if data, exists := rs.Data[key]; exists && data.Expire == mili {
		delete(rs.Data, key)
		fmt.Println("Removed expired key:", key)
	}
}
