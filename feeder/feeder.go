package feeder

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/canthefason/irc-k/client"
	"github.com/canthefason/irc-k/common"
	"github.com/canthefason/irc-k/config"
	"github.com/canthefason/r2dq"
	"gopkg.in/redis.v2"
)

var (
	redisConn      *redis.Client
	conn           *client.Connection
	ErrConnNotInit = errors.New("connection not initialized")
	quit           chan os.Signal
	channels       []string
	queue          *r2dq.Queue
	botName        string
)

const BOT_COUNT = "botcount"

func initialize() {
	redisConn = common.MustGetRedis()
	quit = make(chan os.Signal)
	channels = make([]string, 0)
	queue = common.MustGetQueue()
}

// four different sets
// 1. user - subscribed channels
// 2. server - waiting channels
// 3. server - connected channels (why do i need this?)
// 4. server - channel messages
func Run() {
	connect()
	defer Close()

	go func() {
		for {
			connectToChannel()
		}
	}()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
}

// close iterates over connected channels and adds them to waiting channel list
// for further connections
func Close() {
	go redisConn.Close()
	go func() {
		defer queue.Close()
		gracefulShutdown()
	}()
}

func gracefulShutdown() {
	for _, channel := range channels {
		if err := queue.Queue(channel); err != nil {
			log.Printf("Critical: channel %s could not be requeued: %s", channel, err)
		}
	}
}

func connect() {
	initialize()
	conn = new(client.Connection)
	botName = prepareBotName()

	if err := conn.Connect(botName); err != nil {
		panic(err)
	}
}

func connectToChannel() {
	// get a channel from waiting list
	channel, err := queue.Dequeue()
	if err != nil {
		panic(err)
	}

	// try to join channel
	if err := conn.Join(channel); err != nil {
		log.Printf("An error occurred while joining channel: %s", err)
		queue.NAck(channel)
		return
	}

	log.Printf("%s connected to channel: %s", botName, channel)

	go handleMessages(conn)

	channels = append(channels, channel)
	queue.Ack(channel)
}

func prepareBotName() string {
	res := redisConn.Incr(common.KeyWithPrefix(BOT_COUNT))
	if res.Err() != nil {
		panic(res.Err())
	}

	return fmt.Sprintf("%s-%d", config.Conf.IRC.BotName, res.Val())
}

func handleMessages(conn *client.Connection) {
	for {
		select {
		case m := <-conn.MsgChan:
			if err := common.Send(m); err != nil {
				log.Printf("An error occurred while sending message: %s", err)
			}
		case <-quit:
			return
		}

	}

}
