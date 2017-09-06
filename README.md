# gophirc
A simple IRC bot framework written from scratch, in Go.

## Description
Event based IRC framework.

## Framework managed events 
* Manages server `PING` requests (not `CTCP PING`)
* Registers on first `NOTICE *`
* Identifies on `RPL_WELCOME` (event 001)
* Joins the received invites & sends a greeting to the channel
* Logs if the bot gets kicked from a channel

## Features
* Multiple per event callbacks
* State & general logging
* Graceful exit handled either by a `SIGINT` (Ctrl-C) or a `CTCP QUIT`
* Parses a user from an IRC formatted `nick!user@host` to a `User{}`
* Config implements a basic checking on values
* Already implemented basic commands - `JOIN`, `PART`, `PRIVMSG`, `NOTICE`, `KICK`, `INVITE`, `MODE`, CTCP commands
* Many *(?)* more

## To do
- [ ] Add defaults
- [x] Parse CTCP messages into events
- [x] Add more commands: `MODE`, `KICK`, etc.
- [ ] Add regex matching for nicknames, channels