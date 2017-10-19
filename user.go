package gophirc

import (
	"fmt"
	"regexp"
	"strings"
)

// User stores the nick, user, and host of a user parsed from an event.
type User struct {
	Nick string
	User string
	Host string
}

// String returns "nick!user@host".
func (u *User) String() string {
	if u == nil {
		return ""
	}
	return fmt.Sprintf("%s!%s@%s", u.Nick, u.User, u.Host)
}

// ParseUser splits a "nick!user@host" raw user into a User struct,
// containing the nickname, username, and hostname.
func ParseUser(user string) (*User, bool) {
	if user[0] == ':' {
		user = user[1:]
	}

	pattern := regexp.MustCompile(
		`\A[a-zA-Z_\-\[\]\\^{}|][a-zA-Z0-9_\-\[\]\\^{}|.` + "`" +
			`]*![a-zA-Z0-9_\-\[\]\\^{}|.~` + "`" +
			`]+@[a-zA-Z0-9_\-\[\]\\^{}|.:~` + "`" + `]+\z`)
	if ok := pattern.MatchString(user); !ok {
		return nil, false
	}

	nb := strings.Index(user, "!")
	ub := strings.Index(user, "@")
	if nb == -1 || ub == -1 || nb > ub {
		return nil, false
	}

	return &User{
		Nick: user[:nb],
		User: user[nb+1 : ub],
		Host: user[ub+1:],
	}, true
}
