package gophirc

import (
	"fmt"
	"strings"

	"github.com/vlad-s/gophirc/config"
)

type User struct {
	Nick string
	User string
	Host string
}

func (u User) String() string {
	return fmt.Sprintf("%s!%s@%s", u.Nick, u.User, u.Host)
}

func (u User) IsAdmin() bool {
	for _, v := range config.Get().Admins {
		if v == u.Nick {
			return true
		}
	}
	return false
}

func ParseUser(u string) (*User, bool) {
	if u[0] == ':' {
		u = u[1:]
	}
	nb := strings.Index(u, "!")
	ub := strings.Index(u, "@")
	if nb == -1 || ub == -1 {
		return nil, false
	}
	return &User{
		Nick: u[:nb],
		User: u[nb+1 : ub],
		Host: u[ub+1:],
	}, true
}
