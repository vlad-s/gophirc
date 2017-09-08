package gophirc

import (
	"fmt"
	"strings"
)

type User struct {
	Nick string
	User string
	Host string
}

func (u User) String() string {
	return fmt.Sprintf("%s!%s@%s", u.Nick, u.User, u.Host)
}

// ParseUser splits a "nick!user@host" raw user into a User struct, containing Nick, User, Host fields.
func ParseUser(user string) (*User, bool) {
	if user[0] == ':' {
		user = user[1:]
	}
	nb := strings.Index(user, "!")
	ub := strings.Index(user, "@")
	if nb == -1 || ub == -1 {
		return nil, false
	}
	return &User{
		Nick: user[:nb],
		User: user[nb+1 : ub],
		Host: user[ub+1:],
	}, true
}
