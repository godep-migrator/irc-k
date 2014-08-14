package client

import "errors"

var (
	ErrTimeout       = errors.New("connection timeout")
	ErrChannelNotSet = errors.New("channel not set")
	ErrInternal      = errors.New("internal error")
	ErrNotConnected  = errors.New("not connected")
)
