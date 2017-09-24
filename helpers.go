package gophirc

import "github.com/vlad-s/gophirc/logger"

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

func logStates(i *IRC) {
	var c uint8
	for {
		select {
		case <-i.State.Connected:
			c++
			logger.Log.Infoln("Successfully connected to server")
		case <-i.State.Registered:
			c++
			logger.Log.Infoln("Successfully registered on network")
		case <-i.State.Identified:
			c++
			logger.Log.Infoln("Successfully identified to Nickserv")
		}
		if c == 3 {
			break
		}
	}
}
