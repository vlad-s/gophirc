package gophirc

import "testing"

func TestIsCTCP(t *testing.T) {
	tests := []struct {
		message  string
		expected bool
	}{
		{"sample message", false},
		{"\001ACTION something\001", true},
		{"\x01VERSION\x01", true},
	}
	for _, test := range tests {
		t.Run(test.message, func(t *testing.T) {
			actual := IsCTCP(test.message)
			if actual != test.expected {
				t.Errorf("%q: expected %v, got %v instead.", test.message, test.expected, actual)
			}
		})
	}
}

func TestIsChannel(t *testing.T) {
	tests := []struct {
		channel  string
		expected bool
	}{
		{"#test", true},
		{"test", false},
	}
	for _, test := range tests {
		t.Run(test.channel, func(t *testing.T) {
			actual := IsChannel(test.channel)
			if actual != test.expected {
				t.Errorf("%q: expected %v, got %v instead.", test.channel, test.expected, actual)
			}
		})
	}
}
