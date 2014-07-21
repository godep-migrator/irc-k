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
	quit     chan bool
	nickname *string
	c        *irc.Conn
	connMap  map[string]*Connection
	connRes  chan *Connection
	// channelMap map[string]map[string]struct{}
)

const (
	CONN_TIMEOUT = 5 * time.Second
)

func init() {
	quit = make(chan bool)
	connRes = make(chan *Connection)
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
	Err      error
}

func (c *Connection) sendMessage(m *Message) {
	channel := fmt.Sprintf("#%s", m.Channel)
	c.IrcConn.Join(channel)
	c.IrcConn.Privmsg(channel, m.Body)
}

type Message struct {
	Nickname  string    `json:"nickname" binding:"required"`
	Body      string    `json:"body" binding:"required"`
	Channel   string    `json:"channel" binding:"required"`
	Timestamp time.Time `json:"timestamp"`
}

type Response struct {
	Success bool  `json:"response"`
	Error   error `json:"error"`
}

func connectClient(m *Message) {
	conn, ok := connMap[m.Nickname]
	if ok {
		connRes <- conn
	}

	cfg := irc.NewConfig(m.Nickname)
	cfg.SSL = true
	cfg.Server = "irc.freenode.net:7000"
	cfg.NewNick = func(n string) string { return n + "^" }
	c = irc.Client(cfg)

	conn = new(Connection)
	conn.Nickname = m.Nickname
	if err := c.Connect(); err != nil {
		conn.Err = err
		connRes <- conn
	}

	// just for debugging purposes
	fmt.Printf(c.String())
	conn.IrcConn = c
	connMap[m.Nickname] = conn

	connRes <- conn
}

func prepareErrResponse(err error) Response {
	fmt.Printf("Error occurred: %s", err)
	return Response{Success: false, Error: err}
}

func prepareConnection(m *Message, r render.Render) *Connection {
	go connectClient(m)

	select {
	case conn := <-connRes:
		if conn.Err != nil {
			r.JSON(400, prepareErrResponse(conn.Err))
			return nil
		}
		return conn
	case <-time.After(CONN_TIMEOUT):
		r.JSON(408, prepareErrResponse(errors.New("connection timeout")))
		return nil
	}
}

func sendMessage(_ martini.Params, m Message, r render.Render) {
	conn := prepareConnection(&m, r)

	conn.sendMessage(&m)

	r.JSON(200, Response{true, nil})
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
