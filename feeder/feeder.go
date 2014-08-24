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
	redisConn      *redis.Client
	conn           *client.Connection
	ErrConnNotInit = errors.New("connection not initialized")
	quit           chan os.Signal
	channels       []string
	queue          *r2dq.Queue
	botName        string
	opened         bool
)

const BOT_COUNT = "botcount"

func initialize() {
	redisConn = common.MustGetRedis()
	quit = make(chan os.Signal)
	channels = make([]string, 0)
	queue = common.MustGetQueue()
	opened = true
}

// four different sets
// 1. user - subscribed channels
// 2. server - waiting channels
// 3. server - connected channels (why do i need this?)
// 4. server - channel messages
func Run(i *common.IrcConf) {
	connect(i)
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
	opened = false
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		redisConn.Close()
	}()

	go func() {
		defer wg.Done()
		defer queue.Close()
		gracefulShutdown()
	}()

	wg.Wait()
}

func gracefulShutdown() {
	for _, channel := range channels {
		if err := queue.Queue(channel); err != nil {
			log.Printf("Critical: channel %s could not be requeued: %s", channel, err)
		}
	}
}

func connect(i *common.IrcConf) {
	initialize()
	conn = new(client.Connection)
	conn.Server = i.Server
	conn.Nickname = botName
	botName = prepareBotName(i.BotName)

	if err := conn.Connect(); err != nil {
		panic(err)
	}
}

func connectToChannel() {
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

func prepareBotName(botname string) string {
	res := redisConn.Incr(common.KeyWithPrefix(BOT_COUNT))
	if res.Err() != nil {
		panic(res.Err())
	}

	return fmt.Sprintf("%s-%d", botname, res.Val())
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
