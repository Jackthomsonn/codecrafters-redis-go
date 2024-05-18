package store

type RedisStore struct {
	Data map[RedisObject]RedisObject
}

type RedisObject struct {
	Key   string
	Value string
}

func NewRedisStore() *RedisStore {
	return &RedisStore{Data: make(map[RedisObject]RedisObject)}
}

func (rs *RedisStore) Set(key, value string) {
	rs.Data[RedisObject{Key: key}] = RedisObject{Value: value}
}

func (rs *RedisStore) Get(key string) string {
	return rs.Data[RedisObject{Key: key}].Value
}
