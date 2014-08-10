package common

import (

	"github.com/canthefason/irc-k/config"
	"gopkg.in/redis.v2"
)

const (
	WAITING_CHANNEL_KEY   = "waiting-channel"
	FAILED_CHANNEL_KEY    = "failed-channel"
	CONNECTED_CHANNEL_KEY = "connected-channel"
)

var (
	redisConn    *redis.Client
)

func init() {
	MustInitRedis()
}

func MustInitRedis() {
	redisConf := config.Conf.Redis
	redisConn = redis.NewTCPClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisConf.Server, redisConf.Port),
		Password: "",
		DB:       int64(redisConf.DB),
	})
}

func MustGetRedis() *redis.Client {
	return redisConn
}


func Close() {
	redisConn.Close()
}
}
