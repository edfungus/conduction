package storage

import (
	"errors"
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
	Addr           string
	MaxConnections int
	DatabaseIndex  int
}

// NewRedisStorage returns a new Redis client
func NewRedisStorage(config *RedisStorageConfig) *RedisStorage {
	redisPool := &redigo.Pool{
		MaxIdle:     config.MaxConnections,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redigo.Conn, error) {
			c, err := redigo.Dial("tcp", config.Addr)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("SELECT", config.DatabaseIndex); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
	}
	rs := &RedisStorage{redisPool}
	return rs
}

// Put stores a key-value pair
func (rs *RedisStorage) Put(key string, destination *Destination) error {
	c := rs.pool.Get()
	defer c.Close()

	if destination == nil {
		return errors.New("desination pointer is nil")
	}

	redisValue := &redis.Value{
		Method: destination.Method,
		Path:   destination.Path,
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

// Get retreives the value of a key-value pair
func (rs *RedisStorage) Get(key string) (*Destination, error) {
	c := rs.pool.Get()
	defer c.Close()

	value, err := redigo.Bytes(c.Do("GET", key))
	if err != nil {
		return nil, err
	}
	redisValue := &redis.Value{}
	if err = proto.Unmarshal(value, redisValue); err != nil {
		return nil, err
	}
	return &Destination{
		Method: redisValue.Method,
		Path:   redisValue.Path,
	}, nil
}

// Delete removes a key-value pair
func (rs *RedisStorage) Delete(key string) error {
	c := rs.pool.Get()
	defer c.Close()

	_, err := c.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}

// Close ends the redis session
func (rs *RedisStorage) Close() error {
	err := rs.pool.Close()
	if err != nil {
		return err
	}
	return nil
}
