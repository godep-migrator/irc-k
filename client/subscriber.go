package client

import (
	"github.com/canthefason/irc-k/common"
	"gopkg.in/redis.v2"
)

type Subscriber struct {
	redisConn *redis.Client
}

func NewSubscriber() *Subscriber {
	s := new(Subscriber)
	s.redisConn = common.NewRedis()

	return s
}

func (s *Subscriber) Subscribe(channel string) error {
	if channel == "" {
		return ErrChannelNotSet
	}

	// add channel to members channels
	err := s.redisConn.PubSub().Subscribe(common.KeyWithPrefix(channel))
	if err != nil {
		return err
	}

	// add channel to global channels
	response := s.redisConn.SAdd(common.KeyWithPrefix(common.REQ_CHANNELS_KEY), channel)
	if response.Err() != nil {
		return response.Err()
	}

	if response.Val() == 0 {
		return nil
	}

	if err := common.MustGetQueue().Queue(channel); err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) Close() error {
	return s.redisConn.Close()
}

func removePrefix(channel string) string {
	return strings.Replace(channel, common.PREFIX+":", "", 1)
}
