package main

import (
	"errors"
	"fmt"
	"time"

	irc "github.com/fluffle/goirc/client"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
)

var (
	quit       chan bool
	nickname   *string
	c          *irc.Conn
	connMap    map[string]*Connection
	connRes    chan error
	ErrTimeout = errors.New("connection timeout")
	// channelMap map[string]map[string]struct{}
)

const (
	CONN_TIMEOUT = 5 * time.Second
)

func init() {
	quit = make(chan bool)
	connRes = make(chan error)
	// set up a goroutine to read commands from stdin
	connMap = make(map[string]*Connection)
}

func main() {

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/sendMessage", binding.Json(Message{}), sendMessage)

	m.Run()

}

type Connection struct {
	Nickname string
	IrcConn  *irc.Conn
}

func (c *Connection) sendMessage(m *Message) {
	channel := fmt.Sprintf("#%s", m.Channel)
	c.IrcConn.Join(channel)
	c.IrcConn.Privmsg(channel, m.Body)
}

func (c *Connection) connect(conn *irc.Conn) error {
	c.IrcConn = conn

	go func() {
		if err := c.IrcConn.Connect(); err != nil {
			connRes <- err
			return
		}
		// just for debugging purposes
		fmt.Printf(conn.String())
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

type Message struct {
	Nickname  string    `json:"nickname" binding:"required"`
	Body      string    `json:"body" binding:"required"`
	Channel   string    `json:"channel" binding:"required"`
	Timestamp time.Time `json:"timestamp"`
}

type Response struct {
	Success bool   `json:"response"`
	Error   string `json:"error"`
}

func connectClient(m *Message) (*Connection, error) {
	conn, ok := connMap[m.Nickname]
	if ok {
		return conn, nil
	}

	conn = new(Connection)
	conn.Nickname = m.Nickname

	cfg := irc.NewConfig(m.Nickname)
	cfg.SSL = true
	cfg.Server = "irc.freenode.net:7000"
	cfg.NewNick = func(n string) string { return n + "^" }
	err := conn.connect(irc.Client(cfg))
	if err != nil {
		return nil, err
	}

	connMap[m.Nickname] = conn

	return conn, nil
}

func prepareErrResponse(err error) Response {
	fmt.Printf("Error occurred: %s", err)
	return Response{Success: false, Error: err.Error()}
}

func sendMessage(_ martini.Params, m Message, r render.Render) {
	conn, err := connectClient(&m)
	if err != nil {
		status := 400
		if err == ErrTimeout {
			status = 408
		}
		r.JSON(status, prepareErrResponse(err))
		return
	}

	conn.sendMessage(&m)

	r.JSON(200, Response{true, ""})
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
