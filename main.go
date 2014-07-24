package main

import (
	"github.com/canthefason/irc-k/client"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
)

var (
	connMap map[string]*client.Connection
	// channelMap map[string]map[string]struct{}
)

func init() {
	// set up a goroutine to read commands from stdin
	connMap = make(map[string]*client.Connection)
}

func main() {
	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/sendMessage", binding.Json(client.Message{}), sendMessage)

	m.Run()
}

type Response struct {
	Success bool   `json:"response"`
	Error   string `json:"error"`
}

func fail(r render.Render, err error) {
	var status int

	switch err {
	case client.ErrTimeout:
		status = 408
	default:
		status = 400
	}

	r.JSON(status, Response{Success: false, Error: err.Error()})
}

func sendMessage(_ martini.Params, m client.Message, r render.Render) {
	conn, err := Connect(m.Nickname)
	if err != nil {
		fail(r, err)
		return
	}

	if err := conn.SendMessage(&m); err != nil {
		fail(r, err)
	}

	r.JSON(200, Response{true, ""})
}

func Connect(nickname string) (*client.Connection, error) {
	conn, ok := connMap[nickname]
	if ok {
		return conn, nil
	}

	conn = new(client.Connection)
	conn.Nickname = nickname
	if err := conn.Connect(nickname); err != nil {
		return nil, err
	}

	connMap[nickname] = conn

	return conn, nil
}
