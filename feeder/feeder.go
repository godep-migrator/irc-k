// Package feeder creates a feeder for joining channels,
// and publishes received messages to channel subscribers.
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
	"github.com/canthefason/r2dq"
	redis "gopkg.in/redis.v2"
)

var (
	ErrConnNotInit = errors.New("connection not initialized")

	redisConn *redis.Client
	conn      *client.Connection
	quit      chan os.Signal
	channels  []string
	queue     *r2dq.Queue
	botName   string
	// used for getting joined channels
	joinChan   chan string
	closeQueue chan bool
)

// Used for storing bot count in redis
const BOT_COUNT = "botcount"

func initialize(r *common.RedisConf) {
	redisConn = common.NewRedis(r)
	quit = make(chan os.Signal)
	channels = make([]string, 0)
	joinChan = make(chan string)
	closeQueue = make(chan bool)
	queue = common.MustGetQueue()
}

// Run initializes irc connection via bots, and joins queued
// channels
func Run(i *common.IrcConf, r *common.RedisConf) {
	connect(i, r)
	defer Close()

	go connectToChannel()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
}

// Close iterates over connected channels and adds them to waiting channel list
// for further connections.
func Close() {
	defer queue.Close()
	defer redisConn.Close()
	gracefulShutdown()

	close(quit)
}

func gracefulShutdown() {
	queue.StopDequeue()
	<-closeQueue
	// first close redis connection to prevent further channel consuming
	for _, channel := range channels {
		if err := queue.Queue(channel); err != nil {
			log.Printf("Critical: channel %s can not be requeued: %s", channel, err)
		}
	}
}

func connect(i *common.IrcConf, r *common.RedisConf) {
	initialize(r)
	conn = client.NewConnection()
	conn.Server = i.Server
	conn.Nickname = botName
	botName = prepareBotName(i.BotName)

	if err := conn.Connect(); err != nil {
		panic(err)
	}
}

func connectToChannel() {
	go handleMessages(conn)

	for {
		// get a channel from waiting list
		channel, err := queue.Dequeue()
		if err == r2dq.ErrConnClosed {
			closeQueue <- true
			return
		}

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
		go func() { joinChan <- channel }()

		channels = append(channels, channel)
		queue.Ack(channel)
	}
}

func prepareBotName(botname string) string {
	res := redisConn.Incr(common.KeyWithPrefix(BOT_COUNT))
	if res.Err() != nil {
		panic(res.Err())
	}

	return fmt.Sprintf("%s-%d", botname, res.Val())
}

func handleMessages(conn *client.Connection) {
	for m := range conn.MsgChan {
		if err := common.Send(m); err != nil {
			log.Printf("An error occurred while sending message: %s", err)
		}
	}
}
