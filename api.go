package main

import (
	"errors"
	"fmt"

	"github.com/canthefason/irc-k/client"
	"github.com/canthefason/irc-k/common"
	"github.com/canthefason/irc-k/config"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"gopkg.in/validator.v1"
)

var (
	ErrNotSet  = errors.New("not set")
	ErrUnknown = errors.New("unknown error")
	connMap    map[string]*client.Connection
)

func init() {
	// set up a goroutine to read commands from stdin
	connMap = make(map[string]*client.Connection)
}

func main() {
	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/sendMessage", binding.Json(MessageRequest{}), sendMessage)
	m.Post("/join", binding.Json(ChannelRequest{}), join)

	m.Run()
}

type Response struct {
	Success bool     `json:"response"`
	Errors  []string `json:"errors"`
}

func NewResponse(success bool, errors ...error) Response {
	r := Response{Success: success, Errors: make([]string, 0)}
	for _, error := range errors {
		r.Errors = append(r.Errors, error.Error())
	}

	return r
}

func fail(r render.Render, errors ...error) {
	var status int

	if len(errors) == 0 {
		r.JSON(400, NewResponse(false, ErrUnknown))
		return
	}

	switch errors[0] {
	case client.ErrTimeout:
		status = 408
	case client.ErrInternal:
		status = 500
	default:
		status = 400
	}

	r.JSON(status, NewResponse(false, errors...))
}

func success(r render.Render) {
	r.JSON(200, NewResponse(true))
}

type MessageRequest struct {
	Nickname string `json:"nickname"  binding:"required" validate:"nonzero" `
	Body     string `json:"body" binding:"required" validate:"nonzero"`
	Channel  string `json:"channel" binding:"required" validate:"nonzero"`
}

type ChannelRequest struct {
	Name     string `json:"name" binding:"required" validate:"nonzero"`
	Nickname string `json:"nickname" binding:"required" validate:"nonzero"`
}

func (mr *MessageRequest) mapToMessage() *common.Message {
	m := new(common.Message)
	m.Nickname = mr.Nickname
	m.Body = mr.Body
	m.Channel = mr.Channel

	return m
}

func mapValidatorError(err error) error {
	switch err {
	case validator.ErrZeroValue:
		return ErrNotSet
	default:
		return ErrUnknown
	}
}

func parseValidatorErrors(errs map[string][]error) []error {
	errors := make([]error, 0, len(errs))
	for key, fieldErrs := range errs {
		for _, err := range fieldErrs {
			errors = append(errors, fmt.Errorf("%s %s", key, mapValidatorError(err)))
		}
	}

	return errors
}

func sendMessage(_ martini.Params, mr MessageRequest, r render.Render) {
	valid, errs := validator.Validate(mr)
	if !valid {
		errors := parseValidatorErrors(errs)
		fail(r, errors...)
		return
	}

	conn, err := Connect(mr.Nickname)
	if err != nil {
		fail(r, err)
		return
	}

	m := mr.mapToMessage()
	if err := conn.SendMessage(m); err != nil {
		fail(r, err)
		return
	}

	success(r)
}

func join(_ martini.Params, cr ChannelRequest, r render.Render) {
	valid, errs := validator.Validate(cr)
	if !valid {
		errors := parseValidatorErrors(errs)
		fail(r, errors...)
		return
	}

	c := cr.mapToChannel()
	if err := c.Join(); err != nil {
		fail(r, err)
	}

	success(r)
}

func Connect(nickname string) (*client.Connection, error) {
	conn, ok := connMap[nickname]
	if ok {
		return conn, nil
	}

	conn = new(client.Connection)
	conn.Nickname = nickname
	conn.Server = config.Conf.IRC.Server
	if err := conn.Connect(); err != nil {
		return nil, err
	}

	connMap[nickname] = conn

	return conn, nil
}
