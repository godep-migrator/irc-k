package feeder

import (
	"errors"
	"fmt"
	"log"

	"github.com/canthefason/irc-k/client"
	"github.com/koding/redis"
)

var (
	ErrChannelNotSet     = errors.New("channel not set")
	ErrNicknameNotSet    = errors.New("nickname not set")
	ErrAlreadySubscribed = errors.New("already subscribed")
	redisConn            *redis.RedisSession
)

type Channel struct {
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
}

func init() {
	initRedisConn()
}

func initRedisConn() error {
	var err error
	// TODO add redis settings to config file
	redisConn, err = redis.NewRedisSession(&redis.RedisConf{Server: "localhost:6379", DB: 3})
	if err != nil {
		log.Fatal("Could not connect to redis: %s", err)
	}

	redisConn.SetPrefix("irc-k")

	return nil
}

func run() {
	conn := new(client.Connection)
	// TODO config file
	if err := conn.Connect("koding-bot"); err != nil {
		log.Fatal("Error occurred in initialization: %s", err)
	}
}

// func main() {
// 	conn := new(client.Connection)
// 	// TODO config file
// 	if err := client.Connect("koding-bot"); err != nil {
// 		log.Fatal("Error occurred in initialization: %s", err)
// 	}

// 	http.HandleFunc("/join", joinChannel)
// }

// four different sets
// 1. user - subscribed channels
// 2. server - not connected channels
// 3. server - connected channels
// 4. server - channel messages
func initChannels() {
	for {

	}
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

func (c *Channel) join() error {
	if err := c.validate(); err != nil {
		return err
	}

	if err := c.addUserChannel(); err != nil {
		if err == ErrAlreadySubscribed {
			return nil
		}
	}

	if err := c.addNewChannel(); err != nil {

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
	return nil
}

func prepareUserChannelKey(nickname string) string {
	return fmt.Sprintf("user-channel:%s", nickname)
}
