package common

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/canthefason/r2dq"
	"gopkg.in/redis.v2"
)

const (
	REQ_CHANNELS_KEY = "requested-channels"
	PREFIX           = "irc-k"
)

var (
	redisConn        *redis.Client
	waitingQueue     *r2dq.Queue
	ErrChannelNotSet = errors.New("channel not set")
)

type RedisConf struct {
	Server string
	Port   string
	DB     int
	Prefix string
}

func Initialize(r *RedisConf) *redis.Client {
	MustInitRedis(r)
	MustInitQueue(r)

	return redisConn
}

func MustInitRedis(r *RedisConf) {
	redisConn = NewRedis(r)
}

func NewRedis(r *RedisConf) *redis.Client {
	return redis.NewTCPClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", r.Server, r.Port),
		Password: "",
		DB:       int64(r.DB),
	})
}

func MustGetRedis() *redis.Client {
	return redisConn
}

func MustInitQueue(r *RedisConf) {
	redisAddr := fmt.Sprintf("%s:%s", r.Server, r.Port)
	waitingQueue = r2dq.NewQueue(redisAddr, r.DB, PREFIX)
}

func MustGetQueue() *r2dq.Queue {
	return waitingQueue
}

func Close() {
	redisConn.Close()
	waitingQueue.Close()
}

func KeyWithPrefix(key string) string {
	return fmt.Sprintf("%s:%s", PREFIX, key)
}

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
