package storage

import (
	"time"

	"github.com/edfungus/conduction/storage/redis"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/golang/protobuf/proto"
)

// RedisStorage is the Redis version of storage
type RedisStorage struct {
	pool *redigo.Pool
}

// RedisStorageConfig is the config to make RedisStorage
type RedisStorageConfig struct {
	addr           string
	maxConnections int
}

// NewRedisStorage returns a new Redis client
func NewRedisStorage(config RedisStorageConfig) *RedisStorage {
	redisPool := &redigo.Pool{
		MaxIdle:     config.maxConnections,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redigo.Conn, error) { return redigo.Dial("tcp", config.addr) },
	}
	rs := &RedisStorage{redisPool}
	return rs
}

// Put stores a key-value pair
func (rs *RedisStorage) Put(key string, destination *Destination) error {
	c := rs.pool.Get()
	defer c.Close()

	redisValue := &redis.Value{
		Method: destination.method,
		Path:   destination.path,
	}

	value, err := proto.Marshal(redisValue)
	if err != nil {
		return err
	}
	_, err = c.Do("SET", key, value)
	if err != nil {
		return err
	}
	return nil
}

// Get retreives the value of a keyvalue pair
func (rs *RedisStorage) Get(key string) (*Destination, error) {
	c := rs.pool.Get()
	defer c.Close()

	value, err := redigo.Bytes(c.Do("GET", key))
	redisValue := &redis.Value{}
	if err = proto.Unmarshal(value, redisValue); err != nil {
		return nil, err
	}
	return &Destination{
		method: redisValue.Method,
		path:   redisValue.Path,
	}, nil
}

// Close ends the redis session
func (rs *RedisStorage) Close() error {
	err := rs.pool.Close()
	if err != nil {
		return err
	}
	return nil
}
