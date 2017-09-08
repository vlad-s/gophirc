![](https://img.shields.io/badge/version-0.1-orange.svg?style=flat-square)

# gophirc
A simple IRC bot framework written from scratch, in Go.

## Description
Event based IRC framework.

## Warning
The API might break anytime.

## Framework managed events 
* Manages server `PING` requests (not `CTCP PING`)
* Registers on first `NOTICE *`
* Identifies on `RPL_WELCOME` (event 001)
* Joins the received invites & sends a greeting to the channel
* Logs if the bot gets kicked from a channel

## Features
* Capability to connect to multiple servers
* Multiple per event callbacks
* State & general logging
* Graceful exit handled either by a `SIGINT` (Ctrl-C)
* Parses a user from an IRC formatted `nick!user@host` to a `User{}`
* Config implements a basic checking on values
* Already implemented basic commands - `JOIN`, `PART`, `PRIVMSG`, `NOTICE`, `KICK`, `INVITE`, `MODE`, CTCP commands
* Many *(?)* more

## Events & callbacks
An event follows the next schema:
```go
type Event struct {
    Raw       string   // the whole raw line
    Code      string   // the reply code
    Source    string   // the source of the event (server, user)
    Arguments []string // the arguments after the event code

    User    *User  // if we can parse a user from the source, add the parsed user here
    Message string // if it's a PRIVMSG, add the message here
}
```
You can set callbacks for, technically, all events - numeric reply codes (e.g. "001", "900", etc.) or alpha codes (e.g. "NOTICE", "INVITE", etc.).

Note: CTCP events will have the code set to the corresponding CTCP action, not PRIVMSG.

The framework already binds callbacks for:
* 001 - to identify with NickServ
* 900 - to join the channels specified in config
* NOTICE - only the first event, in order to register with the network
* INVITE - joins the channel & greets

## Examples
Setting a simple bot:
```go
package main

import (
    "github.com/vlad-s/gophirc"
    "github.com/vlad-s/gophirc/config"
)

func main() {
    config.Parse("config.json")
    irc := gophirc.New()
    irc.Connect()
    irc.Loop()
}

```
_Note: error handling remains an exercise for the reader_

Setting up a callback to respond to a CTCP VERSION:
```go
irc.AddEventCallback("VERSION", func(e *gophirc.Event) {
    irc.Notice(e.User.Nick, "\001My own Go bot!\001")
})
```

Setting up a callback to check for custom messages:
```go
irc.AddEventCallback("PRIVMSG", func(e *gophirc.Event) {
    replyTo := e.Arguments[0]
    message := strings.Join(e.Arguments[1:], " ")[1:]

    switch message {
    case "shrug":
        irc.PrivMsg(replyTo, `¯\_(ツ)_/¯`)
    }
})
```

For more examples on usage, please see [gophircbot](https://github.com/vlad-s/gophircbot).

## To do
- [ ] Add defaults
  - [x] Nickname, Username, Realname
- [x] Parse CTCP messages into events
- [x] Add more commands: `MODE`, `KICK`, etc.
- [ ] Add regex matching for nicknames, channels
- [x] Connect Multiple servers