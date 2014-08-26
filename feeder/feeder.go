// Package feeder creates a feeder for joining channels,
// and publishes received messages to channel subscribers.
package feeder

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
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
	opened    bool
	closeChan chan struct{}
)

// Used for storing bot count in redis
const BOT_COUNT = "botcount"

func initialize() {
	redisConn = common.MustGetRedis()
	quit = make(chan os.Signal)
	closeChan = make(chan struct{})
	channels = make([]string, 0)
	queue = common.MustGetQueue()
	opened = true
}

// Run initializes irc connection via bots, and joins queued
// channels
func Run(i *common.IrcConf) {
	connect(i)
	defer Close()

	go connectToChannel()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
	case <-closeChan:
	}
}

// Close iterates over connected channels and adds them to waiting channel list
// for further connections.
func Close() {
	opened = false
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		redisConn.Close()
	}()

	go func() {
		defer wg.Done()
		gracefulShutdown()
		queue.Close()
	}()

	wg.Wait()

	closeChan <- struct{}{}
}

func gracefulShutdown() {
	for _, channel := range channels {
		if err := queue.Queue(channel); err != nil {
			log.Printf("Critical: channel %s can not be requeued: %s", channel, err)
		}
	}
}

func connect(i *common.IrcConf) {
	initialize()
	conn = client.NewConnection()
	conn.Server = i.Server
	conn.Nickname = botName
	botName = prepareBotName(i.BotName)

	if err := conn.Connect(); err != nil {
		panic(err)
	}
}

func connectToChannel() {
	for {
		// get a channel from waiting list
		channel, err := queue.Dequeue()
		if err != nil {
			if !opened {
				return
			}
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
