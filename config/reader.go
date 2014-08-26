// Package config provides configuration data for main package
// For overriding redis host and port settings, REDIS_HOST and
// REDIS_POST environment variables are used
package config

import (
	"log"
	"os"

	"code.google.com/p/gcfg"
	"github.com/canthefason/irc-k/common"
)

type Config struct {
	IRC   common.IrcConf
	Redis common.RedisConf
}

// Exposed root config
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
