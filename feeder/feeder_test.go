package feeder

import "testing"

func NewChannel() *Channel {
	c := new(Channel)
	c.Name = "canthefason-test"
	c.Nickname = "canthefason"

	return c
}

func TestUserJoinValidation(t *testing.T) {
	c := new(Channel)
	err := c.join()
	if err != ErrChannelNotSet {
		t.Error("channel name validation error")
	}

	c.Name = "canthefason-test"
	err = c.join()
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
		t.Errorf("could not add channel: %s", err)
		t.FailNow()
	}

	members, err := redisConn.GetSetMembers(key)
	if err != nil {
		t.Error("could not get channels")
		t.FailNow()
	}

	if len(members) == 0 {
		t.Error("could not get channels")
		t.FailNow()
	}

	result, err := redisConn.String(members[0])
	if err != nil {
		t.Errorf("could not cast channel name to string: %s", err)
		t.FailNow()
	}

	if result != "canthefason-test" {
		t.Errorf("wrong channel name: %s", members[0])
		t.FailNow()
	}

	err = c.addUserChannel()
	if err != ErrAlreadySubscribed {
		t.Error("trying to subscribe to an already joined channel")
	}
}
