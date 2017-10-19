package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/vlad-s/gophirc/logger"
)

// Server contains the information used to a server, namely the address
// and the port, along with the client's details, such as nickname, username, realname,
// NickServ password, channels to join, hardcoded admins & ignored users.
type Server struct {
	Address string `json:"address"`
	Port    uint16 `json:"port"`

	Nickname string `json:"nickname"`
	Username string `json:"username"`
	Realname string `json:"realname"`

	NickservPassword string `json:"nickserv_password"`

	Channels []string `json:"channels"`
	Admins   []string `json:"admins"`
	Ignore   []string `json:"ignore"`
}

// Config dictates the way the config file should be arranged.
type Config struct {
	Servers map[string]*Server `json:"servers"`

	Debug bool `json:"debug"`
}

// Check implements basic error checking and provides some default values for the user.
func (c *Config) Check() error {
	logger.Log.Infoln("Checking the config for errors")

	for name, server := range c.Servers {
		if server.Address == "" {
			return fmt.Errorf("%s: Server address not specified", name)
		}

		if server.Port == 0 {
			return fmt.Errorf("%s: Server port not specified", name)
		}

		if server.Nickname == "" {
			server.Nickname = "gophirc"
		}

		if len(server.Nickname) > 0 && len(server.Nickname) < 3 {
			return fmt.Errorf("%s: Nickname is too short", name)
		}

		if server.Username == "" {
			server.Username = "gophirc"
		}

		if server.Realname == "" {
			server.Realname = "gophirc"
		}
	}

	return nil
}

var conf *Config

// Parse reads and parses the config from the specified path.
func Parse(path string) (*Config, error) {
	logger.Log.Infof("Reading %q", path)

	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "Error opening the config file")
	}
	defer f.Close()

	conf = new(Config)
	err = json.NewDecoder(f).Decode(conf)
	if err != nil {
		return nil, errors.Wrap(err, "Error decoding the config file")
	}
	return conf, nil
}

// Get returns the parsed config, or a new Config{} if it's nil.
func Get() *Config {
	if conf == nil {
		conf = &Config{
			Servers: make(map[string]*Server),
		}
	}
	return conf
}
