package common

import (
	"encoding/json"
	"fmt"

	"github.com/canthefason/irc-k/config"
	"github.com/canthefason/r2dq"
	"gopkg.in/redis.v2"
)

const (
	REQ_CHANNELS_KEY = "requested-channels"
	PREFIX           = "irc-k"
)

var (
	redisConn    *redis.Client
	waitingQueue *r2dq.Queue
)

func init() {
	MustInitRedis()
	MustInitQueue()
}

func MustInitRedis() {
	redisConn = NewRedis()
}

func NewRedis() *redis.Client {
	redisConf := config.Conf.Redis
	return redis.NewTCPClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisConf.Server, redisConf.Port),
		Password: "",
		DB:       int64(redisConf.DB),
	})
}

func MustGetRedis() *redis.Client {
	return redisConn
}

func MustInitQueue() {
	redisConf := config.Conf.Redis
	redisAddr := fmt.Sprintf("%s:%s", redisConf.Server, redisConf.Port)
	waitingQueue = r2dq.NewQueue(redisAddr, redisConf.DB, PREFIX)
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

	res := redisConn.Publish(KeyWithPrefix(m.Channel), string(data))

	return res.Err()
}
