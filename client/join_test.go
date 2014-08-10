package client

import (
	"testing"

	"github.com/canthefason/irc-k/common"
)

func NewChannel() *Channel {
	c := new(Channel)
	c.Name = "canthefason-test"
	c.Nickname = "canthefason"

	return c
}

func TestUserJoinValidation(t *testing.T) {
	c := new(Channel)
	err := c.Join()
	if err != ErrChannelNotSet {
		t.Error("channel name validation error")
	}

	c.Name = "canthefason-test"
	err = c.Join()
	if err != ErrNicknameNotSet {
		t.Error("nickname validation error")
	}
}

func TestUserJoinedChannels(t *testing.T) {
	c := NewChannel()
	err := c.addUserChannel()
	key := prepareUserChannelKey(c.Nickname)
	defer func() {
		redisConn.Del(key)
	}()

	if err != nil {
		t.Errorf("Expected nil but got %s", err)
		t.FailNow()
	}

	res := redisConn.SMembers(key)
	if res.Err() != nil {
		t.Error("Expected nil but got %s", res.Err())
		t.FailNow()
	}

	members := res.Val()
	if len(members) == 0 {
		t.Error("Expected channel list with one channel but got empty list")
		t.FailNow()
	}

	if members[0] != "canthefason-test" {
		t.Errorf("Expected %s but got %s", "canthefason-test", members[0])
		t.FailNow()
	}

	err = c.addUserChannel()
	if err != ErrAlreadySubscribed {
		t.Errorf("Expected %s error but got %s", ErrAlreadySubscribed, err)
	}
}

func TestAddNewChannel(t *testing.T) {
	c := NewChannel()
	defer func() {
		redisConn.Del(common.KeyWithPrefix(common.REQ_CHANNELS_KEY))
		redisConn.Del(prepareUserChannelKey("canthefason"))
		common.Close()
	}()

	err := c.addNewChannel()
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

	err = c.addNewChannel()
	if err != ErrChannelJoined {
		t.Errorf("Expected %s error but got %s", ErrChannelJoined, err)
	}

}
