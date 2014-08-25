// Package client has two responsibilities:
//
// 1. Establishes irc connections via Connection
//
// 2. Subscribes to channel messages via Subscriber
//
// Connection serves as a simple irc client wrapper for sending/receiving
// messages and joining public channels. It does not
// support any other irc operations.
//
// Subscriber can be considered as an irc channel feeder.
// After subscribing a channel, user is notified for further messages.
package client

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/canthefason/irc-k/common"
	irc "github.com/fluffle/goirc/client"
)

const (
	CONN_TIMEOUT = 5 * time.Second
)

// Connection holds the required connection data for client
type Connection struct {
	// preferred user nickname
	Nickname string

	// received messages are piped into this channel
	MsgChan chan common.Message

	// freenode server definition. it must be defined as host:port
	Server string

	// irc connection
	ircConn *irc.Conn

	// connection result errors are piped
	connRes chan error

	// disconnected event publishes state through this channel
	quit chan bool
}

// NewConnection creates connection instance with channels
func NewConnection() *Connection {
	c := new(Connection)
	c.connRes = make(chan error)
	c.quit = make(chan bool)

	return c
}

// prepareChannel forms channel name in format #(channel-name)
func prepareChannel(channel string) string {
	return fmt.Sprintf("#%s", channel)
}

// SendMessage validates message, joins to given channel and sends message.
// Before sending any messages connection must be established first
func (c *Connection) SendMessage(m *common.Message) error {
	if err := m.Validate(); err != nil {
		return err
	}

	if err := c.Join(m.Channel); err != nil {
		return err
	}

	channel := prepareChannel(m.Channel)
	c.ircConn.Privmsg(channel, m.Body)

	return nil
}

// Connect creates irc connection and register event handlers.
// If connection times out, it returns timeout error
func (c *Connection) Connect() error {

	cfg := irc.NewConfig(c.Nickname)
	cfg.SSL = true
	cfg.Server = c.Server
	cfg.NewNick = func(n string) string { return n + "^" }
	c.ircConn = irc.Client(cfg)

	go func() {
		if err := c.ircConn.Connect(); err != nil {
			log.Printf("an error occurred: %s \n", err)
			c.connRes <- ErrInternal
			return
		}
		// just for debugging purposes
		fmt.Printf(c.ircConn.String())
		c.connRes <- nil
	}()

	select {
	case err := <-c.connRes:
		if err != nil {
			return err
		}
	case <-time.After(CONN_TIMEOUT):
		c.ircConn.Quit()
		return ErrTimeout
	}
	// TODO instead of sleeping, it needs to handle connected event
	time.Sleep(time.Second * 15)
	c.registerHandlers()

	return nil
}

// Join connects user to given channel if irc connection is established
func (c *Connection) Join(channelName string) error {
	if channelName == "" {
		return ErrChannelNotSet
	}

	channel := prepareChannel(channelName)
	if c.ircConn == nil {
		return ErrNotConnected
	}

	c.ircConn.Join(channel)

	return nil
}

// registerHandler registers user to connected, privmsg and disconnected
// irc events. When more operators needed, handlers must be registered here.
func (c *Connection) registerHandlers() {
	c.MsgChan = make(chan common.Message, 0)
	c.ircConn.HandleFunc("connected",
		func(conn *irc.Conn, line *irc.Line) {
			// connected
		})

	c.ircConn.HandleFunc("disconnected",
		//TODO handle disconnection
		func(conn *irc.Conn, line *irc.Line) {
			c.quit <- true
		})

	c.ircConn.HandleFunc("privmsg",
		func(conn *irc.Conn, line *irc.Line) {
			channel := line.Args[0]
			if strings.IndexRune(channel, '#') == 0 {
				channel = strings.Replace(channel, "#", "", 1)
			}
			m := common.Message{
				Nickname: line.Nick,
				Body:     line.Args[1],
				Channel:  channel,
			}

			c.MsgChan <- m
		},
	)
}
