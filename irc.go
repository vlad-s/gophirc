package gophirc

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/vlad-s/gophirc/config"
	"github.com/vlad-s/gophirc/logger"
)

type State struct {
	Connected chan struct{}

	registered bool
	Registered chan struct{}

	Identified chan struct{}
}

type Event struct {
	Raw       string
	Code      string
	Source    string
	Arguments []string

	User    *User
	Message string
}

type IRC struct {
	conn   net.Conn
	Server config.Server

	State  State
	Events map[string][]func(*Event)

	raw  chan string
	quit chan struct{}
}

func (irc *IRC) Connect() error {
	dest := fmt.Sprintf("%s:%d", irc.Server.Address, irc.Server.Port)
	c, err := net.DialTimeout("tcp", dest, 5*time.Second)
	if err != nil {
		return errors.Wrap(err, "Error dialing the host")
	}
	irc.conn = c
	return nil
}

func (irc *IRC) Disconnect(s string) {
	fmt.Fprint(irc.conn, fmt.Sprintf("QUIT :%s\r\n", s))
	irc.conn.Close()
}

func (irc *IRC) Loop() error {
	var gracefulExit bool

	go func() {
		for {
			select {
			case <-irc.quit:
				gracefulExit = true
				irc.Disconnect("SIGINT")
				return
			case s := <-irc.raw:
				if config.Get().Debug {
					logger.Log.Debugf("Raw:\t%q", s)
				}
			}
		}
	}()

	s := bufio.NewScanner(irc.conn)
	for s.Scan() {
		go irc.getRaw(s.Text())
	}

	if gracefulExit {
		return nil
	}
	return errors.Wrap(s.Err(), "Error while looping")
}

func (irc *IRC) AddEventCallback(code string, cb func(*Event)) *IRC {
	irc.Events[code] = append(irc.Events[code], cb)
	return irc
}

func (irc *IRC) parseToEvent(raw string) (event *Event, ok bool) {
	irc.raw <- raw
	event = &Event{Raw: raw}
	if raw[0] != ':' {
		return
	}

	raw = raw[1:]
	split := strings.Split(raw, " ")

	event.Source = split[0]
	event.Code = split[1]
	event.Arguments = split[2:]

	if u, ok := ParseUser(event.Source); ok {
		event.User = u
	}

	if event.Code == "PRIVMSG" {
		message := strings.Join(event.Arguments[1:], " ")[1:]
		if IsCTCP(message) {
			message = strings.Trim(message, "\001")
			message_args := strings.Split(message, " ")
			event.Code = message_args[0]
			event.Arguments = message_args[1:]
		}
		event.Message = message
	}

	return event, true
}

func (irc *IRC) getRaw(raw string) {
	e, ok := irc.parseToEvent(raw)
	if !ok {
		split := strings.Split(e.Raw, " ")
		if split[0] == "PING" {
			irc.pong(split[1])
		}
	}

	for k, v := range irc.Events {
		if k != e.Code {
			continue
		}
		for _, f := range v {
			f(e)
		}
	}

	switch e.Code {
	case "404":
		logger.Log.WithField("channel", e.Arguments[0]).Warnln("Can't send to channel")
	case "474":
		logger.Log.WithField("channel", e.Arguments[0]).Warnln("Can't join channel")
	case "KICK":
		if e.Arguments[1] == config.Get().Nickname {
			logger.Log.WithFields(logger.Fields(map[string]interface{}{
				"user": e.User.Nick, "channel": e.Arguments[0],
			})).Warnln("We got kicked from a channel")
		}
	}
}

func (irc *IRC) addBasicCallbacks() {
	irc.AddEventCallback("NOTICE", func(e *Event) {
		if e.Arguments[0] == "*" && !irc.State.registered {
			irc.State.Connected <- struct{}{}
			irc.Register()
		}
	}).AddEventCallback("001", func(e *Event) {
		irc.Identify()
	}).AddEventCallback("900", func(e *Event) {
		irc.State.Identified <- struct{}{}
		for _, v := range irc.Server.Channels {
			irc.Join(v)
		}
	}).AddEventCallback("INVITE", func(e *Event) {
		channel := e.Arguments[1][1:]
		irc.Join(channel)
		irc.PrivMsg(channel, fmt.Sprintf("Hi %s, %s invited me here.", channel, e.User.Nick))
	})
}

func (irc *IRC) Quit() {
	irc.quit <- struct{}{}
}

func New() *IRC {
	conf := config.Get()
	logger.Log.WithFields(logger.Fields(map[string]interface{}{
		"server": conf.Server.Address, "port": conf.Server.Port,
	})).Infoln("Connecting to server")

	i := &IRC{
		Server: conf.Server,

		State: State{
			Connected:  make(chan struct{}),
			Registered: make(chan struct{}),
			Identified: make(chan struct{}),
		},
		Events: make(map[string][]func(*Event)),

		raw:  make(chan string),
		quit: make(chan struct{}, 1),
	}

	i.addBasicCallbacks()
	go logStates(i)

	return i
}
