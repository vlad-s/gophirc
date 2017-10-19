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

// State keeps track of the framework's states, as the name implies.
type State struct {
	Registered   bool
	Disconnected struct {
		Value     bool
		Requested bool
	}
}

// Event contains the raw event received from the server along with the parsed data.
// The framework parses CTCP messages as event, changing the Code to the CTCP action received.
// In case we receive an event from a user (and not the server), we parse it and store it into
// the `User` variable.
type Event struct {
	Raw       string   // Raw line received from the server
	Code      string   // Event code received or parsed from a CTCP message
	Source    string   // Source of the event, user or server
	Arguments []string // Arguments of the event

	User    *User  // If the source is a user, parse it & store it
	Message string // If the event is a PRIVMSG, store the message here
	ReplyTo string // Store the user or the channel to reply to
}

// IRC is the main structure containing the connection, server, state, event callbacks, etc.
type IRC struct {
	conn   net.Conn
	Server *config.Server

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

	irc.State.Disconnected = struct {
		Value     bool
		Requested bool
	}{Value: false, Requested: false}

	return nil
}

// Disconnect sends a QUIT command to the server, and closes the connection.
func (irc *IRC) Disconnect(s string) {
	fmt.Fprint(irc.conn, fmt.Sprintf("QUIT :%s\r\n", s))

	irc.State.Disconnected = struct {
		Value     bool
		Requested bool
	}{Value: true, Requested: true}

	//irc.conn.Close()
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

	irc.State.Disconnected = struct {
		Value     bool
		Requested bool
	}{Value: true, Requested: false}

	logger.Log.Errorln(errors.Wrap(s.Err(), "Error while looping"))

	err := irc.Connect()
	if err != nil {
		logger.Log.Errorln(errors.Wrap(err, "Can't (re)connect to server"))
	}
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
			messageArgs := strings.Split(message, " ")
			event.Code = messageArgs[0]
			event.Arguments = messageArgs[1:]
		}
		event.Message = strings.TrimSpace(message)

		if len(event.Arguments) > 0 {
			if IsChannel(event.Arguments[0]) {
				event.ReplyTo = event.Arguments[0]
			} else if event.Arguments[0] == irc.Server.Nickname {
				event.ReplyTo = event.User.Nick
			}
		}
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
			return
		}
	}

	if irc.IsIgnored(e.User) {
		return
	}

	if e.User != nil && e.User.Nick == irc.Server.Nickname {
		// our own event
		//return
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

			if strings.Contains(e.Raw, "*** Looking up") && e.User == nil {
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
		logger.Log.Infoln("Successfully connected to server")
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
	logger.Log.Infoln("Successfully identified to Nickserv")
	for _, v := range irc.Server.Channels {
		logger.Log.Infof("Joining channel %q", v)
		irc.Join(v)
	}
}

// Quit provides a wrapper to sending a value into the `quit` channel.
func (irc *IRC) Quit() {
	irc.quit <- struct{}{}
}

// IsAdmin returns whether or not the specified user is an admin.
func (irc *IRC) IsAdmin(u *User) bool {
	if u == nil {
		return false
	}
	for _, v := range irc.Server.Admins {
		if u.Nick == v {
			return true
		}
	}
	return false
}

// IsIgnored returns whether or not the specified user is ignored.
func (irc *IRC) IsIgnored(u *User) bool {
	if u == nil {
		return false
	}
	for _, v := range irc.Server.Ignore {
		if u.Nick == v {
			return true
		}
	}
	return false
}

// New returns a pointer to a new IRC struct using the server & wait group specified.
func New(server *config.Server, wg *sync.WaitGroup) *IRC {
	logger.Log.WithFields(logger.Fields(map[string]interface{}{
		"server": server.Address, "port": server.Port,
	})).Infoln("Generating new server connection")

	i := &IRC{
		Server: server,

		Events: make(map[string][]func(*Event)),

		Waiter: wg,

		raw:  make(chan string),
		quit: make(chan struct{}, 1),
	}

	i.addBasicCallbacks()

	return i
}
