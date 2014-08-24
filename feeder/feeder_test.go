package feeder

import (
	"os"
	"testing"

	"github.com/canthefason/irc-k/common"
)

func tearUp() {
	conf := &common.IrcConf{
		Server:  "irc.freenode.net:7000",
		BotName: "momo",
	}
	rConf := &common.RedisConf{
		Server: "localhost",
		Port:   "6379",
		DB:     3,
		Prefix: "irc-test",
	}

	// TODO later on handle this via github.com/danryan/env
	if env := os.Getenv("REDIS_HOST"); env != "" {
		rConf.Server = env
	}
	if env := os.Getenv("REDIS_PORT"); env != "" {
		rConf.Port = os.Getenv("REDIS_PORT")
	}

	common.Initialize(rConf)
	connect(conf)
}

func tearDown() {
	redisConn.Del(common.KeyWithPrefix(BOT_COUNT))
	queue.Purge()
}

func TestPrepareBotName(t *testing.T) {
	tearUp()
	initialize()
	defer tearDown()
	botName := prepareBotName("momo")
	expectedBotName := "momo-2"
	if botName != expectedBotName {
		t.Errorf("Expected %s but got %s", expectedBotName, botName)
	}

}

func TestInitChannels(t *testing.T) {
	tearUp()
	defer tearDown()

	queue.Queue("test-channel")
	connectToChannel()

	if len(channels) != 1 {
		t.Errorf("Expected 1 but got %d", len(channels))
	}
}

func TestCloseFeeder(t *testing.T) {
	tearUp()
	defer tearDown()
	queue.Queue("test-channel")
	length, err := queue.Len()
	if err != nil {
		t.Errorf("Expected nil but got %s", err)
	}

	if length != 1 {
		t.Errorf("Expected %d but got %d", 1, length)
	}

	connectToChannel()

	length, _ = queue.Len()
	if length != 0 {
		t.Errorf("Expected %d but got %d", 0, length)
	}

	gracefulShutdown()

	length, _ = queue.Len()
	if length != 1 {
		t.Errorf("Expected %d but got %d", 1, length)
	}
}
