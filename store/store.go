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
	Expire int64
}

func NewRedisStore() *RedisStore {
	return &RedisStore{Data: make(map[string]RedisObject)}
}

func (rs *RedisStore) Set(res internal.ParsedResponse) {
	rs.mu.Lock()
	expires := time.Now().Add(time.Duration(res.Mili)*time.Millisecond).UnixNano() / 1e6
	rs.Data[res.Key] = RedisObject{Value: res.Value, Expire: expires}
	rs.mu.Unlock()
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

func RunRemovalCheck(rs *RedisStore) {
	for {
		rs.mu.Lock()
		for key, data := range rs.Data {
			if data.Expire < time.Now().UnixNano()/1e6 {
				delete(rs.Data, key)
				fmt.Println("Removed expired key:", key)
			}
		}
		rs.mu.Unlock()
		time.Sleep(1 * time.Second)
	}
}
