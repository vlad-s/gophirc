package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/vlad-s/gophirc/logger"
)

type Server struct {
	Address string `json:"address"`
	Port    uint16 `json:"port"`

	Nickname string `json:"nickname"`
	Username string `json:"username"`
	Realname string `json:"realname"`

	NickservPassword string `json:"nickserv_password"`

	Channels []string `json:"channels"`
	Admins   []string `json:"admins"`
}

type Config struct {
	Servers map[string]Server `json:"servers"`

	Debug bool `json:"debug"`
}

func (c *Config) Check() error {
	logger.Log.Infoln("Checking the config for errors")

	for name, server := range c.Servers {
		if server.Address == "" {
			return errors.New(fmt.Sprintf("%s: Server address not specified", name))
		}

		if server.Port == 0 {
			return errors.New(fmt.Sprintf("%s: Server port not specified", name))
		}

		if server.Nickname == "" {
			server.Nickname = "gophirc"
		}

		if len(server.Nickname) > 0 && len(server.Nickname) < 3 {
			return errors.New(fmt.Sprintf("%s: Nickname is too short", name))
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

func Parse(s string) (*Config, error) {
	logger.Log.Infof("Reading %q", s)

	f, err := os.Open(s)
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

func Get() *Config {
	if conf == nil {
		return new(Config)
	}
	return conf
}
