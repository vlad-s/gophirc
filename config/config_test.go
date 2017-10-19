package config

import (
	"testing"
)

func TestGet(t *testing.T) {
	conf = nil
	conf := Get()
	if conf == nil || len(conf.Servers) > 0 {
		t.Error("Error getting a new config")
	}

	_, err := Parse("config.json.example")
	if err != nil {
		t.Error("Error getting the parsed config")
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		shouldFail bool
	}{
		{"config.json.example", false},
		{"inexistent.config.json", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conf, err := Parse(test.name)
			if (test.shouldFail && err == nil) || (!test.shouldFail && err != nil) {
				t.Errorf("Config %q - should fail: %v, got err %q\n", test.name, test.shouldFail, err)
			}
			if !test.shouldFail {
				nick := conf.Servers["first"].Nickname
				if nick != "gophirc" {
					t.Errorf("Config %q - wrong nickname; expected \"gophirc\", got %q\n", test.name, nick)
				}
			}
		})
	}
}

func TestConfig_CheckEmpty(t *testing.T) {
	conf = nil
	Get()
	if err := conf.Check(); err != nil {
		t.Error("Error checking empty config")
	}
}

func TestConfig_CheckConf(t *testing.T) {
	Parse("config.json.example")
	if err := conf.Check(); err != nil {
		t.Error("Error checking default config")
	}
	s := conf.Servers["second"]

	if s.Nickname != "my_bot" {
		t.Errorf("Error checking the nickname, got %q, expected %q\n", s.Nickname, "my_bot")
	}

	s.Nickname = ""
	conf.Check()
	if s.Nickname != "gophirc" {
		t.Errorf("Error checking the nickname, got %q, expected %q", s.Nickname, "gophirc")
	}

	s.Nickname = "a"
	if err := conf.Check(); err == nil {
		t.Errorf("Error checking a short username, got %q", s.Nickname)
	}
	s.Nickname = "gophirc"

	if s.Username != "gophirc" {
		t.Errorf("Error checking the username, got %q, expected %q\n", s.Username, "gophirc")
	}

	if s.Realname != "gophirc" {
		t.Errorf("Error checking the realname, got %q, expected %q\n", s.Realname, "gophirc")
	}

	s.Address = ""
	if err := conf.Check(); err == nil {
		t.Errorf("Error checking the server address, got %q", s.Address)
	}
	s.Address = "irc.other.server.tld"

	s.Port = 0
	if err := conf.Check(); err == nil {
		t.Errorf("Error checking the server port, got %q", s.Port)
	}
	s.Port = 6667
}
