package feeder

import (
	"encoding/json"

	"github.com/canthefason/irc-k/common"
)

func Send(m common.Message) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	res := redisConn.Publish(m.Channel, string(data))

	return res.Err()
}
