package common

import "errors"

var (
	ErrBodyNotSet     = errors.New("body not set")
	ErrNicknameNotSet = errors.New("nickname not set")
)

type Message struct {
	Nickname string `json:"nickname"`
	Body     string `json:"body"`
	Channel  string `json:"-"`
}

func (m *Message) Validate() error {
	if m.Body == "" {
		return ErrBodyNotSet
	}

	if m.Nickname == "" {
		return ErrNicknameNotSet
	}

	return nil
}
