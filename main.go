package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	irc "github.com/fluffle/goirc/client"
	"github.com/rcrowley/go-tigertonic"
)

var (
	quit     chan bool
	in       chan string
	nickname *string
	c        *irc.Conn
	connMap  map[string]*irc.Conn
	// channelMap map[string]map[string]struct{}
)

func init() {
	quit = make(chan bool)
	// set up a goroutine to read commands from stdin
	in = make(chan string)
	connMap = make(map[string]*irc.Conn)
}

func main() {

	// go connect("koding-bot", channels)

	mux := tigertonic.NewTrieServeMux()
	// mux.Handle("POST", "/connect", tigertonic.Timed(tigertonic.Marshaled(connect), "connect", nil))
	// mux.Handle("POST", "/join", tigertonic.Timed(tigertonic.Marshaled(join), "join", nil))
	mux.Handle("POST", "/sendMessage", tigertonic.Timed(tigertonic.Marshaled(sendMessage), "sendMessage", nil))

	tigertonic.NewServer(":9000", tigertonic.Logged(mux, nil)).ListenAndServe()

}

// func connect(u *url.URL, h http.Header, conn *Connection) (int, http.Header, *Response, error) {

// 	registerHandlers(c)

// 	if err := c.Connect(); err != nil {
// 		return http.StatusBadRequest, nil, &Response{false, err}, nil
// 	} else {
// 		fmt.Printf(c.String())
// 	}

// 	connMap[conn.Nickname] = c

// 	return http.StatusOK, nil, &Response{true, nil}, nil
// }

type Message struct {
	Nickname  string    `json:"nickname"`
	Body      string    `json:"body"`
	Channel   string    `json:"channel"`
	Timestamp time.Time `json:"timestamp"`
}

type Response struct {
	Response bool  `json:"response"`
	Error    error `json:"error"`
}

func connect(m *Message) (*irc.Conn, error) {
	fmt.Println("hu")
	conn, ok := connMap[m.Nickname]
	if ok {
		return conn, nil
	}

	fmt.Println("ho")

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

	// time.After(500 * time.Millisecond)

	connMap[m.Nickname] = c

	return c, nil
}

func sendMessage(_ *url.URL, _ http.Header, m *Message) (int, http.Header, *Response, error) {
	fmt.Println("mey")
	conn, err := connect(m)
	if err != nil {
		return http.StatusBadRequest, nil, &Response{Response: false, Error: err}, nil
	}

	fmt.Println("hey")

	channel := fmt.Sprintf("#%s", m.Channel)
	conn.Join(channel)

	// time.Sleep()

	fmt.Println("body", m.Body)
	conn.Privmsg(channel, m.Body)

	return http.StatusOK, nil, &Response{true, nil}, nil
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
