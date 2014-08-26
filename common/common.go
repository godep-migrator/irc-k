// Package common provides redis and channel queue connections
// Common is used by client, config and feeder packages
package common

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/canthefason/r2dq"
	"gopkg.in/redis.v2"
)

const (
	// used for storing all channel names in a set
	REQ_CHANNELS_KEY = "requested-channels"

	// redis key prefix
	PREFIX = "irc-k"
)

var (
	redisConn        *redis.Client
	waitingQueue     *r2dq.Queue
	ErrChannelNotSet = errors.New("channel not set")
	ErrRedisNotInit  = errors.New("redis not initialized")
	ErrQueueNotInit  = errors.New("queue not initialized")
)

// RedisConf holds redis connection data
type RedisConf struct {
	Server string
	Port   string
	DB     int
	Prefix string
}

// IrcConf holds irc connection data
type IrcConf struct {
	// Server name is set as hostname:port
	Server string
	// Stores botname to be used by feeder
	BotName string
}

// Initialize initilizes redis and queue connections
func Initialize(r *RedisConf) *redis.Client {
	MustInitRedis(r)
	MustInitQueue(r)

	return redisConn
}

// MustInitRedis initialize global redis connection
func MustInitRedis(r *RedisConf) {
	redisConn = NewRedis(r)
}

// NewRedis creates a new redis connection
func NewRedis(r *RedisConf) *redis.Client {
	return redis.NewTCPClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", r.Server, r.Port),
		Password: "",
		DB:       int64(r.DB),
	})
}

// MustGetRedis returns global redis connection. It panics when
// redis connection is not yet initialized
func MustGetRedis() *redis.Client {
	if redisConn == nil {
		panic(ErrRedisNotInit)
	}

	return redisConn
}

// MustInitQueue initializes channel queue
func MustInitQueue(r *RedisConf) {
	redisAddr := fmt.Sprintf("%s:%s", r.Server, r.Port)
	waitingQueue = r2dq.NewQueue(redisAddr, r.DB, PREFIX)
}

// MustGetQueue returns channel queue
func MustGetQueue() *r2dq.Queue {
	if waitingQueue == nil {
		panic(ErrQueueNotInit)
	}

	return waitingQueue
}

// Close closes both redis and queue connections
func Close() {
	redisConn.Close()
	waitingQueue.Close()
}

// KeyWithPrefix appends prefix constant to the given key
func KeyWithPrefix(key string) string {
	return fmt.Sprintf("%s:%s", PREFIX, key)
}

// Send publishes message to the related channel
func Send(m Message) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	if m.Channel == "" {
		return ErrChannelNotSet
	}

	res := redisConn.Publish(KeyWithPrefix(m.Channel), string(data))

	return res.Err()
}
