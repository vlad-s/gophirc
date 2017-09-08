package gophirc

import (
	"fmt"
	"strings"

	"github.com/vlad-s/gophirc/config"
)

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

func (irc *IRC) pong(s string) {
	irc.SendRawf("PONG %s", s)
}

func (irc *IRC) Register() {
	conf := config.Get()
	irc.SendRawf("USER %s 8 * %s", conf.Username, conf.Realname)
	irc.SendRawf("NICK %s", conf.Nickname)

	irc.State.registered = true
	irc.State.Registered <- struct{}{}
}

func (irc *IRC) Identify() {
	ns := config.Get().Server.NickservPassword
	if ns == "" {
		return
	}
	irc.SendRawf("NS IDENTIFY %s", ns)
}

func (irc *IRC) Join(channel string) {
	irc.SendRawf("JOIN %s", channel)
}

func (irc *IRC) Part(channel string) {
	irc.SendRawf("PART %s", channel)
}

func (irc *IRC) PrivMsg(replyTo, message string) {
	irc.SendRawf("PRIVMSG %s :%s", replyTo, message)
}

func (irc *IRC) Notice(replyTo, message string) {
	irc.SendRawf("NOTICE %s :%s", replyTo, message)
}

func (irc *IRC) Action(replyTo, message string) {
	irc.SendRawf("PRIVMSG %s :\001ACTION %s\001", replyTo, message)
}

func (irc *IRC) CTCP(replyTo, ctcp, message string) {
	irc.Notice(replyTo, fmt.Sprintf("\001%s %s\001", ctcp, message))
}

func (irc *IRC) Kick(channel, nick, message string) {
	if message == "" {
		message = "Requested"
	}
	irc.SendRawf("KICK %s %s :%s", channel, nick, message)
}

func (irc *IRC) Invite(nick, channel string) {
	irc.SendRawf("INVITE %s %s", nick, channel)
}

func (irc *IRC) Mode(channel, mode, nick string) {
	irc.SendRawf("MODE %s %s %s", channel, mode, nick)
}

func (irc *IRC) Ban(channel, nick string) {
	irc.Mode(channel, "+b", nick)
}

func (irc *IRC) Unban(channel, nick string) {
	irc.Mode(channel, "-b", nick)
}

func (irc *IRC) KickBan(channel, nick string) {
	irc.Ban(channel, nick)
	irc.Kick(channel, nick, "beep boop i press buttons")
}

func (irc *IRC) Nick(nick string) {
	irc.SendRawf("NICK %s", nick)
}
