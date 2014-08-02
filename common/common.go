package common

import (
	"errors"
	"log"

	"github.com/canthefason/irc-k/config"
	"github.com/koding/redis"
)

const (
	WAITING_CHANNEL_KEY   = "waiting-channel"
	FAILED_CHANNEL_KEY    = "failed-channel"
	CONNECTED_CHANNEL_KEY = "connected-channel"
)

var redisConn *redis.RedisSession

func init() {
	MustInitRedis()
}

func MustInitRedis() {
	var err error
	redisConf := config.Conf.Redis
	redisConn, err = redis.NewRedisSession(&redis.RedisConf{Server: redisConf.Server, DB: redisConf.DB})
	if err != nil {
		log.Fatal("Could not connect to redis: %s", err)
	}

	redisConn.SetPrefix(redisConf.Prefix)
}

func MustGetRedis() *redis.RedisSession {
	if redisConn == nil {
		panic(errors.New("redis is not initialized"))
	}

	return redisConn
}
