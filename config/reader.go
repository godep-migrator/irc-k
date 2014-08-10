package config

import (
	"log"

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
}
