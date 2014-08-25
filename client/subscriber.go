package client

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/canthefason/irc-k/common"
	"gopkg.in/redis.v2"
)

// Subscriber holds the receive channel for fetching
// messages of subscribed channels
type Subscriber struct {
	// message reception channel
	Rcv chan common.Message

	redisConn *redis.Client
	ps        *redis.PubSub
	quit      bool
}

// NewSubscriber creates redis and pubsub connections and opens
// receive channel
func NewSubscriber(r *common.RedisConf) *Subscriber {
	s := new(Subscriber)

	s.redisConn = common.Initialize(r)
	s.Rcv = make(chan common.Message, 0)
	s.ps = s.redisConn.PubSub()

	return s
}

// Subscribe used for subscribing a user to given channel messages.
func (s *Subscriber) Subscribe(channel string) error {
	if channel == "" {
		return ErrChannelNotSet
	}

	// subscribe user to given channel for receiving channel messages
	err := s.ps.Subscribe(common.KeyWithPrefix(channel))
	if err != nil {
		return err
	}

	// add subscribed channel to requested channels set. it is used for
	// preventing duplicate channel connections
	response := s.redisConn.SAdd(common.KeyWithPrefix(common.REQ_CHANNELS_KEY), channel)
	if response.Err() != nil {
		return response.Err()
	}

	if response.Val() == 0 {
		log.Printf("bot is already connected to channel: %s", channel)
		return nil
	}

	// queue channel name for feeder connection. this queue is consumed by feeder workers.
	if err := common.MustGetQueue().Queue(channel); err != nil {
		return err
	}

	return nil
}

// Listen starts listening channel messages in blocking manner.
func (s *Subscriber) Listen() {
	for {
		res, err := s.ps.Receive()
		if err != nil {
			// when connection is closed, it returns err.
			// if quit flag is set, this err message is ignored.
			if s.quit {
				return
			}
			panic(err)
		}

		// res also returns subscription events.
		// we are ignoring them and just listening messages here
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

// Close ends pub/sub and redis connections and sets quit flag to true
func (s *Subscriber) Close() error {
	s.quit = true
	s.ps.Close()
	return s.redisConn.Close()
}

// removePrefix is used for removing redis cache key prefix.
func removePrefix(channel string) string {
	return strings.Replace(channel, common.PREFIX+":", "", 1)
}
