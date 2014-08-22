package client

import (
	"testing"
	"time"

	"github.com/canthefason/irc-k/common"
)

func tearUp() *Subscriber {
	conf := &common.RedisConf{
		Server: "localhost",
		Port:   6379,
		DB:     3,
		Prefix: "irc-test",
	}
	s := NewSubscriber(conf)
}

func tearDown(s *Subscriber) {
	go func() {
		s.redisConn.Del(common.KeyWithPrefix(common.REQ_CHANNELS_KEY))
		s.Close()
	}()
	go func() {
		common.MustGetQueue().Purge()
		// common.Close()
	}()
}

func TestUserSubscribeValidation(t *testing.T) {
	s := tearUp()
	defer tearDown(s)

	err := s.Subscribe("")
	if err != ErrChannelNotSet {
		t.Error("Expected nil but got %s", err)
	}

	err = s.Subscribe("canthefason-test")
	if err != nil {
		t.Error("Expected nil but got %s", err)
	}
}

func TestAddNewChannel(t *testing.T) {
	s := tearUp()
	defer tearDown(s)

	err := s.Subscribe("canthefason-test")
	if err != nil {
		t.Errorf("Expected nil but got %s", err)
		t.FailNow()
	}

	pkg, err := common.MustGetQueue().Dequeue()
	if err != nil {
		t.Errorf("Expected nil but got %s", err)
		t.FailNow()
	}

	if pkg != "canthefason-test" {
		t.Errorf("Expected %s but got %s", "canthefason-test", pkg)
		t.FailNow()
	}
}

func TestRemovePrefix(t *testing.T) {
	channelName := common.KeyWithPrefix("labs")
	c := removePrefix(channelName)
	if "labs" != c {
		t.Errorf("Expected %s but got %s", "labs", c)
	}
}

func TestListenChannel(t *testing.T) {
	s := tearUp
	if err := s.Subscribe("muppet-kitchen"); err != nil {
		t.Errorf("Expected nil but got %s", err)
	}
	go s.Listen()

	defer tearDown(s)
	completed := make(chan struct{}, 1)

	m := common.Message{}
	m.Nickname = "swedishchef"
	m.Body = "dudupdudup"
	m.Channel = "muppet-kitchen"

	go func() {
		msg := <-s.Rcv
		if msg.Nickname != m.Nickname {
			t.Errorf("Expected %s as nickname but got %s", m.Nickname, msg.Nickname)
		}

		if msg.Body != m.Body {
			t.Errorf("Expected %s as body but got %s", m.Body, msg.Body)
		}

		if msg.Channel != m.Channel {
			t.Errorf("Expected %s as channel but got %s", m.Channel, msg.Channel)
		}

		completed <- struct{}{}
	}()

	if err := common.Send(m); err != nil {
		t.Error("Expected nil but got %s", err)
	}

	select {
	case <-completed:
	case <-time.After(time.Second * 2):
		t.Error("Expected message but connection timeout")
	}
}
