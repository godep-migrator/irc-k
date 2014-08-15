package client

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/canthefason/irc-k/common"
	"github.com/canthefason/irc-k/config"
	irc "github.com/fluffle/goirc/client"
)

var (
	quit    chan bool
	connRes chan error
)

const (
	CONN_TIMEOUT = 5 * time.Second
)

type Connection struct {
	Nickname string
	MsgChan  chan common.Message

	ircConn *irc.Conn
}

func init() {
	connRes = make(chan error)
	quit = make(chan bool)
}

// prepareChannel appends # sign to channel name
func prepareChannel(channel string) string {
	return fmt.Sprintf("#%s", channel)
}

func (c *Connection) SendMessage(m *common.Message) error {
	if err := m.Validate(); err != nil {
		return err
	}

	if err := c.Join(m.Channel); err != nil {
		return err
	}

	c.ircConn.Privmsg(m.Channel, m.Body)

	return nil
}

func (c *Connection) Connect(nickname string) error {

	cfg := irc.NewConfig(nickname)
	cfg.SSL = true
	cfg.Server = config.Conf.IRC.Server
	cfg.NewNick = func(n string) string { return n + "^" }
	c.ircConn = irc.Client(cfg)

	go func() {
		if err := c.ircConn.Connect(); err != nil {
			log.Printf("an error occurred: %s \n", err)
			connRes <- ErrInternal
			return
		}
		// just for debugging purposes
		fmt.Printf(c.ircConn.String())
		connRes <- nil
	}()

	select {
	case err := <-connRes:
		if err != nil {
			return err
		}
	case <-time.After(CONN_TIMEOUT):
		c.ircConn.Quit()
		return ErrTimeout
	}

	c.registerHandlers()

	return nil
}

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

func (c *Connection) registerHandlers() {
	c.MsgChan = make(chan common.Message, 0)
	c.ircConn.HandleFunc("connected",
		func(conn *irc.Conn, line *irc.Line) {
			// connected
		})

	c.ircConn.HandleFunc("disconnected",
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

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
