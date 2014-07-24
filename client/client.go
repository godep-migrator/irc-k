package client

import (
	"errors"
	"fmt"
	"time"

	irc "github.com/fluffle/goirc/client"
)

var (
	ErrTimeout        = errors.New("connection timeout")
	ErrBodyNotSet     = errors.New("body not set")
	ErrChannelNotSet  = errors.New("channel not set")
	ErrNicknameNotSet = errors.New("nickname not set")
	ErrBadRequest     = errors.New("an error occurred")
	ErrNotConnected   = errors.New("not connected")
	quit              chan bool
	connRes           chan error
)

const (
	CONN_TIMEOUT = 5 * time.Second
)

type Connection struct {
	Nickname string
	IrcConn  *irc.Conn
}

func init() {
	connRes = make(chan error)
	quit = make(chan bool)
}

type Message struct {
	Nickname  string    `json:"nickname" binding:"required"`
	Body      string    `json:"body" binding:"required"`
	Channel   string    `json:"channel" binding:"required"`
	Timestamp time.Time `json:"timestamp"`
}

func (m *Message) validate() error {
	if m.Body == "" {
		return ErrBodyNotSet
	}

	if m.Channel == "" {
		return ErrChannelNotSet
	}

	if m.Nickname == "" {
		return ErrNicknameNotSet
	}

	return nil
}

// prepareChannel appends # sign to channel name
func prepareChannel(channel string) string {
	return fmt.Sprintf("#%s", channel)
}

func (c *Connection) SendMessage(m *Message) error {
	if err := m.validate(); err != nil {
		return err
	}

	channel := prepareChannel(m.Channel)
	if c.IrcConn == nil {
		return ErrNotConnected
	}

	c.IrcConn.Join(channel)
	c.IrcConn.Privmsg(channel, m.Body)

	return nil
}

func (c *Connection) Connect(nickname string) error {

	cfg := irc.NewConfig(nickname)
	cfg.SSL = true
	cfg.Server = "irc.freenode.net:7000" // TODO read it from a config file
	cfg.NewNick = func(n string) string { return n + "^" }
	c.IrcConn = irc.Client(cfg)

	go func() {
		if err := c.IrcConn.Connect(); err != nil {
			fmt.Printf("an error occurred: %s \n", err)
			connRes <- ErrBadRequest
			return
		}
		// just for debugging purposes
		fmt.Printf(c.IrcConn.String())
		connRes <- nil
	}()

	select {
	case err := <-connRes:
		if err != nil {
			return err
		}
	case <-time.After(CONN_TIMEOUT):
		c.IrcConn.Quit()
		return ErrTimeout
	}

	return nil
}

func registerHandlers(c *irc.Conn) {
	c.HandleFunc("connected",
		func(conn *irc.Conn, line *irc.Line) {
			// connected
		})

	c.HandleFunc("disconnected",
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	c.HandleFunc("privmsg",
		func(conn *irc.Conn, line *irc.Line) {
			fmt.Printf("%s : %+v \n", line.Nick, line.Args[1])
		},
	)
}
