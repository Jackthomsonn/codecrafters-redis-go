package store

import (
	"fmt"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal"
)

type RedisStore struct {
	Data []RedisObject
}

type RedisObject struct {
	Key    string
	Value  interface{}
	Expire uint64
}

func NewRedisStore() *RedisStore {
	return &RedisStore{Data: []RedisObject{}}
}

func (rs *RedisStore) Set(res internal.ParsedResponse) {
	rs.Data = append(rs.Data, RedisObject{Key: res.Key, Value: res.Value, Expire: res.Mili})
	go rs.handleRemovalOfExpiredData(res.Mili, res.Key)
}

func (rs *RedisStore) Get(key string) string {
	for _, data := range rs.Data {
		if data.Key == key {
			return data.Value.(string)
		}
	}

	return "$-1\r\n"
}

func (rs *RedisStore) handleRemovalOfExpiredData(mili uint64, key string) {
	fmt.Println("Sleeping for ", mili)
	if mili == 0 {
		return
	}
	time.Sleep(time.Duration(mili) * time.Millisecond)
	for i, data := range rs.Data {
		if data.Key == key {
			rs.Data = remove(rs.Data, i)
		}
	}
}

func remove(s []RedisObject, i int) []RedisObject {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
