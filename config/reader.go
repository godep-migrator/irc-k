package config

import (
	"log"
	"os"

	"code.google.com/p/gcfg"
)

type Config struct {
	IRC struct {
		Server  string
		BotName string
	}
	Redis struct {
		Server string
		Port   string
		DB     int
		Prefix string
	}
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
