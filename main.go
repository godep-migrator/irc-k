package main

import (
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
	connMap  map[string]*irc.Conn
	// channelMap map[string]map[string]struct{}
)

func init() {
	quit = make(chan bool)
	// set up a goroutine to read commands from stdin
	connMap = make(map[string]*irc.Conn)
}

func main() {

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/sendMessage", binding.Json(Message{}), sendMessage)

	m.Run()

}

type Message struct {
	Nickname  string    `json:"nickname" binding:"required"`
	Body      string    `json:"body" binding:"required"`
	Channel   string    `json:"channel" binding:"required"`
	Timestamp time.Time `json:"timestamp"`
}

type Response struct {
	Response bool  `json:"response"`
	Error    error `json:"error"`
}

func connect(m *Message) (*irc.Conn, error) {
	conn, ok := connMap[m.Nickname]
	if ok {
		return conn, nil
	}

	cfg := irc.NewConfig(m.Nickname)
	cfg.SSL = true
	cfg.Server = "irc.freenode.net:7000"
	cfg.NewNick = func(n string) string { return n + "^" }
	c = irc.Client(cfg)

	if err := c.Connect(); err != nil {
		return nil, err
	} else {
		fmt.Printf(c.String())
	}

	connMap[m.Nickname] = c

	return c, nil
}

func sendMessage(_ martini.Params, m Message, r render.Render) {
	conn, err := connect(&m)
	if err != nil {
		r.JSON(400, Response{Response: false, Error: err})
		return
	}

	channel := fmt.Sprintf("#%s", m.Channel)
	conn.Join(channel)

	conn.Privmsg(channel, m.Body)

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
