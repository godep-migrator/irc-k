package client

import (
	"errors"
	"fmt"

	"github.com/canthefason/irc-k/common"
	"gopkg.in/redis.v2"
)

var (
	ErrAlreadySubscribed = errors.New("already subscribed")
	ErrChannelJoined     = errors.New("already joined")
	redisConn            *redis.Client
)

const (
	USER_CHANNEL_KEY = "user-channel"
)

type Channel struct {
	Name     string
	Nickname string
}

func init() {
	redisConn = common.MustGetRedis()
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
	res := redisConn.SAdd(prepareUserChannelKey(c.Nickname), c.Name)
	if res.Err() != nil {
		return res.Err()
	}

	if res.Val() == 0 {
		return ErrAlreadySubscribed
	}

	return nil
}

func (c *Channel) addNewChannel() error {
	// add channel to members channels
	res := redisConn.SAdd(prepareUserChannelKey(c.Nickname), c.Name)
	if res.Err() != nil {
		return res.Err()
	}

	if res.Val() == 0 {
		return ErrChannelJoined
	}

	// add channel to global channels
	response := redisConn.SAdd(common.KeyWithPrefix(common.REQ_CHANNELS_KEY), c.Name)
	if response.Err() != nil {
		return res.Err()
	}

	if response.Val() == 0 {
		return nil
	}

	return nil
}

func prepareUserChannelKey(nickname string) string {
	return fmt.Sprintf("%s:%s", USER_CHANNEL_KEY, nickname)
}
