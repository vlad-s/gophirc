package gophirc

import "github.com/vlad-s/gophirc/logger"

func IsCTCP(s string) bool {
	if s[0] == '\001' && s[len(s)-1:][0] == '\001' {
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
