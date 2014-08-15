package common

import (
	"encoding/json"
	"testing"
	"time"

	"gopkg.in/redis.v2"
)

func tearDown() {
	MustGetQueue().Purge()
}

func TestMessagePublish(t *testing.T) {
	defer tearDown()
	redisSubConn := NewRedis()

	done := make(chan *redis.Message, 1)
	quit := make(chan struct{}, 1)

	ps := redisSubConn.PubSub()
	if err := ps.Subscribe(KeyWithPrefix("electric-mayhem")); err != nil {
		t.Errorf("Expected nil but got %s", err)
	}

	go func() {
		for {
			res, err := ps.Receive()
			if err != nil {
				t.Errorf("Expected nil but got %s", err)
				quit <- struct{}{}
			}

			switch res.(type) {
			case *redis.Message:
				done <- res.(*redis.Message)
			}

		}

	}()

	m := Message{}
	m.Body = "can you picture that?"
	m.Nickname = "kermit"

	err := Send(m)
	if err != ErrChannelNotSet {
		t.Errorf("Expected %s but got %s", ErrChannelNotSet, err)
	}

	m.Channel = "electric-mayhem"
	err = Send(m)
	if err != nil {
		t.Errorf("Expected nil but got %s", err)
	}

	select {
	case res := <-done:
		rm := new(Message)
		err := json.Unmarshal([]byte(res.Payload), &rm)
		if err != nil {
			t.Errorf("Expected nil but got %s", err)
		}

		if rm.Body != m.Body {
			t.Errorf("Expected %s as message body but got %s", m.Body, rm.Body)
		}
		if rm.Nickname != m.Nickname {
			t.Errorf("Expected %s as nickname but got %s", m.Nickname, rm.Nickname)
		}
	case <-time.After(time.Second * 3):
		t.Error("Could not get message")
	case <-quit:
	}
}
