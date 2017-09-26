package gophirc

// IsCTCP returns whether a string is a CTCP action based on the 0x01 flag.
func IsCTCP(s string) bool {
	if s[0] == '\001' && s[len(s)-1:][0] == '\001' {
		return true
	}
	return false
}

// IsChannel returns whether a string is a channel or not.
func IsChannel(s string) bool {
	// todo: add regex for better matching
	if s[0] == '#' {
		return true
	}
	return false
}
