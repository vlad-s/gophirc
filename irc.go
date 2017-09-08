package gophirc

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
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

	Waiter *sync.WaitGroup

	raw  chan string
	quit chan struct{}
}

// Connect tries to connect to the server with the address & port specified in the config.
// It has a 5 second timeout on the dialing.
func (irc *IRC) Connect() error {
	dest := fmt.Sprintf("%s:%d", irc.Server.Address, irc.Server.Port)
	c, err := net.DialTimeout("tcp", dest, 5*time.Second)
	if err != nil {
		return errors.Wrap(err, "Error dialing the host")
	}
	irc.Waiter.Add(1)
	irc.conn = c
	return nil
}

// Disconnect sends a QUIT command to the server, and closes the connection.
func (irc *IRC) Disconnect(s string) {
	fmt.Fprint(irc.conn, fmt.Sprintf("QUIT :%s\r\n", s))
	irc.conn.Close()
	irc.Waiter.Done()
}

// Loop keeps the connection active, getting the raw text from the server.
// It also handles the quitting & debug logging.
func (irc *IRC) Loop() {
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
					logger.Log.Debugf("%s:%d - %q", irc.Server.Address, irc.Server.Port, s)
				}
			}
		}
	}()

	s := bufio.NewScanner(irc.conn)
	for s.Scan() {
		go irc.ReadEvent(s.Text())
	}

	if gracefulExit {
		return
	}

	logger.Log.Errorln(errors.Wrap(s.Err(), "Error while looping"))
}

// AddEventCallback adds a callback function to the Events map on the specified reply code.
func (irc *IRC) AddEventCallback(code string, cb func(*Event)) *IRC {
	irc.Events[code] = append(irc.Events[code], cb)
	return irc
}

// ParseToEvent reads and parses a raw string to an Event struct.
func (irc *IRC) ParseToEvent(raw string) (event *Event, ok bool) {
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
		event.Message = strings.TrimSpace(message)
	}

	return event, true
}

// ReadEvent reads a parsed Event, calls the callbacks defined, and adds some basic logging.
func (irc *IRC) ReadEvent(raw string) {
	e, ok := irc.ParseToEvent(raw)
	if !ok {
		split := strings.Split(e.Raw, " ")
		if split[0] == "PING" {
			irc.pong(split[1])
		}
	}

	for _, callback := range irc.Events[e.Code] {
		callback(e)
	}

	switch e.Code {
	case "404":
		logger.Log.WithField("channel", e.Arguments[0]).Warnln("Can't send to channel")
	case "474":
		logger.Log.WithField("channel", e.Arguments[0]).Warnln("Can't join channel")
	case "KICK":
		if e.Arguments[1] == irc.Server.Nickname {
			logger.Log.WithFields(logger.Fields(map[string]interface{}{
				"user": e.User.Nick, "channel": e.Arguments[0],
			})).Warnln("We got kicked from a channel")
		}
	}
}

func (irc *IRC) addBasicCallbacks() {
	irc.AddEventCallback("NOTICE", func(e *Event) {
		go func(e *Event) {
			message := strings.Join(e.Arguments[1:], " ")

			if e.Arguments[0] == "*" && !irc.State.registered {
				irc.State.Connected <- struct{}{}
				irc.Register()
			}

			if e.User == nil {
				return
			}

			if e.User.Nick == "NickServ" && strings.HasPrefix(message, ":Password accepted") {
				go irc.autojoin(e)
			}
		}(e)
	}).AddEventCallback("001", func(e *Event) {
		irc.Identify()
	}).AddEventCallback("900", func(e *Event) {
		go irc.autojoin(e)
	}).AddEventCallback("INVITE", func(e *Event) {
		go func(e *Event) {
			channel := e.Arguments[1][1:]
			irc.Join(channel)
			irc.PrivMsg(channel, fmt.Sprintf("Hi %s, %s invited me here.", channel, e.User.Nick))
		}(e)
	})
}

func (irc *IRC) autojoin(e *Event) {
	irc.State.Identified <- struct{}{}
	for _, v := range irc.Server.Channels {
		irc.Join(v)
	}
}

// Quit provides a wrapper to sending a value into the `quit` channel.
func (irc *IRC) Quit() {
	irc.quit <- struct{}{}
}

// IsAdmin returns whether or not the specified user is an admin.
func (irc *IRC) IsAdmin(u *User) bool {
	for _, v := range irc.Server.Admins {
		if u.Nick == v {
			return true
		}
	}
	return false
}

// New returns a pointer to a new IRC struct using the server & wait group specified.
func New(server config.Server, wg *sync.WaitGroup) *IRC {
	logger.Log.WithFields(logger.Fields(map[string]interface{}{
		"server": server.Address, "port": server.Port,
	})).Infoln("Connecting to server")

	i := &IRC{
		Server: server,

		State: State{
			Connected:  make(chan struct{}),
			Registered: make(chan struct{}),
			Identified: make(chan struct{}),
		},
		Events: make(map[string][]func(*Event)),

		Waiter: wg,

		raw:  make(chan string),
		quit: make(chan struct{}, 1),
	}

	i.addBasicCallbacks()
	go logStates(i)

	return i
}
