package feeder

import "encoding/json"

type Message struct {
	Body     string `json:"body"`
	Nickname string `json:"nickname"`
	Channel  string `json:"-"`
}

func Send(m Message) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	res := redisConn.Publish(m.Channel, string(data))

	return res.Err()
}
