package client

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/canthefason/irc-k/common"
	"gopkg.in/redis.v2"
)

type Subscriber struct {
	redisConn *redis.Client
	Rcv       chan common.Message
	ps        *redis.PubSub
}

func NewSubscriber(r *common.RedisConf) *Subscriber {
	s := new(Subscriber)

	s.redisConn = common.Initialize(r)
	s.Rcv = make(chan common.Message, 0)
	s.ps = s.redisConn.PubSub()

	return s
}

func (s *Subscriber) Subscribe(channel string) error {
	if channel == "" {
		return ErrChannelNotSet
	}

	// add channel to members channels
	err := s.ps.Subscribe(common.KeyWithPrefix(channel))
	if err != nil {
		return err
	}

	// add channel to global channels
	response := s.redisConn.SAdd(common.KeyWithPrefix(common.REQ_CHANNELS_KEY), channel)
	if response.Err() != nil {
		return response.Err()
	}

	if response.Val() == 0 {
		log.Printf("bot is already connected to channel: %s", channel)
		return nil
	}

	if err := common.MustGetQueue().Queue(channel); err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) Listen() error {
	for {
		res, err := s.ps.Receive()
		if err != nil {
			panic(err)
		}

		switch res.(type) {
		case *redis.Message:
			rm := res.(*redis.Message)
			msg := common.Message{}
			err := json.Unmarshal([]byte(rm.Payload), &msg)
			if err != nil {
				log.Printf("Could not unmarshal received message: %s", err)
				continue
			}
			msg.Channel = removePrefix(rm.Channel)
			s.Rcv <- msg
		}

	}
}

func (s *Subscriber) Close() error {
	return s.redisConn.Close()
}

func removePrefix(channel string) string {
	return strings.Replace(channel, common.PREFIX+":", "", 1)
}
