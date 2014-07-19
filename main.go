package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	irc "github.com/fluffle/goirc/client"
)

var (
	quit chan bool
	in   chan string
)

func init() {
	quit = make(chan bool)
	// set up a goroutine to read commands from stdin
	in = make(chan string)
}

func main() {
	// k := kite.New("irc", "0.0.1")

	cfg := irc.NewConfig("kodingcan")
	cfg.SSL = true
	cfg.Server = "irc.freenode.net:7000"
	cfg.NewNick = func(n string) string { return n + "^" }
	c := irc.Client(cfg)

	registerHandlers(c)

	go readInput()

	// set up a goroutine to do parsey things with the stuff from stdin
	go parseCommand(c)

	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err)
	} else {
		fmt.Printf(c.String())
	}

	<-quit
}

func readInput() {
	con := bufio.NewReader(os.Stdin)
	for {
		s, err := con.ReadString('\n')
		if err != nil {
			// wha?, maybe ctrl-D...
			close(in)
			break
		}
		// no point in sending empty lines down the channel
		if len(s) > 2 {
			in <- s[0 : len(s)-1]
		}
	}
}

func parseCommand(c *irc.Conn) {
	for cmd := range in {
		if cmd[0] == ':' {
			switch idx := strings.Index(cmd, " "); {
			case cmd[1] == 'd':
				fmt.Printf(c.String())
			case cmd[1] == 'f':
				if len(cmd) > 2 && cmd[2] == 'e' {
					// enable flooding
					c.Config().Flood = true
				} else if len(cmd) > 2 && cmd[2] == 'd' {
					// disable flooding
					c.Config().Flood = false
				}
				c.Privmsg("#canthefason-test", cmd[idx+1:len(cmd)])
			case idx == -1:
				continue
			case cmd[1] == 'q':
				c.Quit(cmd[idx+1 : len(cmd)])
				quit <- true
			case cmd[1] == 'j':
				c.Join(cmd[idx+1 : len(cmd)])
			case cmd[1] == 'p':
				c.Part(cmd[idx+1 : len(cmd)])
			}
		} else {
			c.Raw(cmd)
		}
	}
}

func registerHandlers(c *irc.Conn) {
	c.HandleFunc("connected",
		func(conn *irc.Conn, line *irc.Line) { conn.Join("#canthefason-test") })

	c.HandleFunc("disconnected",
		func(conn *irc.Conn, line *irc.Line) { quit <- true })

	c.HandleFunc("privmsg",
		func(conn *irc.Conn, line *irc.Line) {
			fmt.Printf("%s : %+v \n", line.Nick, line.Args[1])
		},
	)
}
