package client

import (
	"errors"
	"fmt"
	"log"

	"github.com/canthefason/irc-k/config"
	"github.com/koding/redis"
)

var (
	ErrAlreadySubscribed = errors.New("already subscribed")
	ErrChannelJoined     = errors.New("already joined")
	redisConn            *redis.RedisSession
)

const (
	USER_CHANNEL_KEY      = "user-channel"
	WAITING_CHANNEL_KEY   = "waiting-channel"
	CONNECTED_CHANNEL_KEY = "connected-channel"
)

type Channel struct {
	Name     string
	Nickname string
}

func init() {
	initRedisConn()
}

func initRedisConn() error {
	var err error
	redisConf := config.Conf.Redis
	redisConn, err = redis.NewRedisSession(&redis.RedisConf{Server: redisConf.Server, DB: redisConf.DB})
	if err != nil {
		log.Fatal("Could not connect to redis: %s", err)
	}

	redisConn.SetPrefix("irc-k")

	return nil
}

func (c *Channel) validate() error {
	if c.Name == "" {
		return ErrChannelNotSet
	}

	if c.Nickname == "" {
		return ErrNicknameNotSet
	}

	return nil
}

func (c *Channel) Join() error {
	if err := c.validate(); err != nil {
		return err
	}

	if err := c.addUserChannel(); err != nil {
		if err == ErrAlreadySubscribed {
			return nil
		}
	}

	if err := c.addNewChannel(); err != nil {
		if err == ErrChannelJoined {
			return nil
		}
	}

	return nil
}

func (c *Channel) addUserChannel() error {
	reply, err := redisConn.AddSetMembers(prepareUserChannelKey(c.Nickname), c.Name)
	if err != nil {
		return err
	}

	if reply == 0 {
		return ErrAlreadySubscribed
	}

	return nil
}

func (c *Channel) addNewChannel() error {
	reply, err := redisConn.IsSetMember(CONNECTED_CHANNEL_KEY, c.Name)
	if err != nil {
		return err
	}

	if reply == 1 {
		return ErrChannelJoined
	}

	reply, err = redisConn.AddSetMembers(WAITING_CHANNEL_KEY, c.Name)
	if err != nil {
		return err
	}

	if reply == 0 {
		return ErrChannelJoined
	}

	return nil
}

func prepareUserChannelKey(nickname string) string {
	return fmt.Sprintf("%s:%s", USER_CHANNEL_KEY, nickname)
}
