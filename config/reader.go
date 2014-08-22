package config

import (
	"log"
	"os"

	"code.google.com/p/gcfg"
	"github.com/canthefason/irc-k/common"
)

type Config struct {
	IRC struct {
		Server  string
		BotName string
	}
	Redis common.RedisConf
}

var Conf *Config

func init() {
	Conf = new(Config)
	if err := gcfg.ReadStringInto(Conf, confStr); err != nil {
		log.Fatalf("Could not initialize config file: %s", err)
		return
	}

	if env := os.Getenv("REDIS_HOST"); env != "" {
		Conf.Redis.Server = env
	}
	if env := os.Getenv("REDIS_PORT"); env != "" {
		Conf.Redis.Port = os.Getenv("REDIS_PORT")
	}
}
