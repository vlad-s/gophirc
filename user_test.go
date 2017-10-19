package gophirc

import (
	"testing"
)

func TestParseUser(t *testing.T) {
	tests := []struct {
		user     string
		expected bool
	}{
		{"NiCkNaM3!uSeRnAmE@server-0o1.073.39ec7k.IP", true},
		{":}o{!I`mAButterfly@this.is.my.vhost", true},
		{"x@y!z", false},
		{"malformed", false},
		{"psycho!~madness@0x00.0x70737963686f", true},
		{"gophirc_test!~gophirc@2a02:2f0d:1a1:c19:581e:ca94:3650:2615", true},
	}
	for _, test := range tests {
		t.Run(test.user, func(t *testing.T) {
			_, actual := ParseUser(test.user)
			if actual != test.expected {
				t.Errorf("User %q - expected %v, got %v\n", test.user, test.expected, actual)
			}
		})
	}
}

func TestUser_String(t *testing.T) {
	tests := []struct {
		user     *User
		expected string
	}{
		{&User{"a", "b", "c"}, "a!b@c"},
		{&User{"foo", "bar", "baz"}, "foo!bar@baz"},
		{&User{"}o{", "I`mAButterfly", "this.is.my.vhost"}, "}o{!I`mAButterfly@this.is.my.vhost"},
		{nil, ""},
	}
	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			if test.user.String() != test.expected {
				t.Errorf("User %+q, expected %q, got %q\n", test.user, test.expected, test.user.String())
			}
		})
	}
}
