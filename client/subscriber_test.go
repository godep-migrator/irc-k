package client

import (
	"testing"

	"github.com/canthefason/irc-k/common"
)

func tearDown(s *Subscriber) {
	go func() {
		s.redisConn.Del(common.KeyWithPrefix(common.REQ_CHANNELS_KEY))
		s.Close()
	}()
	go func() {
		common.MustGetQueue().Purge()
		common.Close()
	}()
}

func TestUserSubscribeValidation(t *testing.T) {
	s := NewSubscriber()
	defer s.Close()

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
	s := NewSubscriber()
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
