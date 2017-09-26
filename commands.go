package gophirc

import (
	"fmt"
	"strings"

	"github.com/vlad-s/gophirc/logger"
)

// SendRaw sends a raw string back to the server, appending a CR LF.
// It automatically strips carriage returns and line feeds from the string.
func (irc *IRC) SendRaw(s string) {
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	irc.raw <- s
	fmt.Fprint(irc.conn, s+"\r\n")
}

func (irc *IRC) SendRawf(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	s = strings.Replace(s, "\r", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	irc.raw <- s
	fmt.Fprint(irc.conn, s+"\r\n")
}

// pong sends a PONG command back to server. Used after receiving a PING.
func (irc *IRC) pong(s string) {
	irc.SendRawf("PONG %s", s)
}

// Register sends the USER and NICK commands to the server, and sets the registered state.
func (irc *IRC) Register() {
	irc.SendRawf("USER %s 8 * %s", irc.Server.Username, irc.Server.Realname)
	irc.SendRawf("NICK %s", irc.Server.Nickname)

	irc.State.registered = true
	logger.Log.Infoln("Successfully registered on network")
}

// Identify sends the NickServ identify command to the server.
func (irc *IRC) Identify() {
	ns := irc.Server.NickservPassword
	if ns == "" {
		return
	}
	irc.SendRawf("NS IDENTIFY %s", ns)
}

// Join sends a JOIN command to the server, requesting to join <channel>.
func (irc *IRC) Join(channel string) {
	irc.SendRawf("JOIN %s", channel)
}

// Part sents a PART command to the server, requesting to part <channel>.
func (irc *IRC) Part(channel string) {
	irc.SendRawf("PART %s", channel)
}

// PrivMsg sends a PRIVMSG command to the server.
func (irc *IRC) PrivMsg(replyTo, message string) {
	irc.SendRawf("PRIVMSG %s :%s", replyTo, message)
}

func (irc *IRC) PrivMsgf(replyTo, format string, args ...interface{}) {
	irc.SendRawf("PRIVMSG %s :%s", replyTo, fmt.Sprintf(format, args...))
}

// Notice sends a NOTICE command to the server.
func (irc *IRC) Notice(replyTo, message string) {
	irc.SendRawf("NOTICE %s :%s", replyTo, message)
}

// Action sends a PRIVMSG command to the server, mimicking the "/me" in IRC clients.
func (irc *IRC) Action(replyTo, message string) {
	irc.SendRawf("PRIVMSG %s :\001ACTION %s\001", replyTo, message)
}

// CTCP sends a notice to the server, with CTCP flag 0x01 appended & prepended.
func (irc *IRC) CTCP(replyTo, ctcp, message string) {
	irc.Notice(replyTo, fmt.Sprintf("\001%s %s\001", ctcp, message))
}

// Kick sends a KICK command to the server, requesting to kick <nick> from <channel> using <message>.
func (irc *IRC) Kick(channel, nick, message string) {
	if message == "" {
		message = "Requested"
	}
	irc.SendRawf("KICK %s %s :%s", channel, nick, message)
}

// Invite sends an INVITE command to the server, requesting to invite <nick> to <channel>.
func (irc *IRC) Invite(nick, channel string) {
	irc.SendRawf("INVITE %s %s", nick, channel)
}

// Mode sends a MODE command to the server, requesting to set/unset mode <mode> on <nick> on <channel>.
func (irc *IRC) Mode(channel, mode, nick string) {
	irc.SendRawf("MODE %s %s %s", channel, mode, nick)
}

// Ban sends a MODE +b command to the server, requesting to ban <nick> on <channel>.
func (irc *IRC) Ban(channel, nick string) {
	irc.Mode(channel, "+b", nick)
}

// Unban sends a MODE -b command to the server, requesting to unban <nick> on <channel>.
func (irc *IRC) Unban(channel, nick string) {
	irc.Mode(channel, "-b", nick)
}

// KickBan bans the user and then kicks him from the channel.
func (irc *IRC) KickBan(channel, nick string) {
	irc.Ban(channel, nick)
	irc.Kick(channel, nick, "beep boop i press buttons")
}

// Nick sends a NICK command to the server, requesting to change the current nick into <nick>.
func (irc *IRC) Nick(nick string) {
	irc.SendRawf("NICK %s", nick)
}
