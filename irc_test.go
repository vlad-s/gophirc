package gophirc

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/vlad-s/gophirc/config"
)

var wg sync.WaitGroup
var irc *IRC

var channel = "#gophirc_test" + strconv.Itoa(time.Now().Second())

func TestNew(t *testing.T) {
	// no error checking, config tests imply this is working
	conf, _ := config.Parse("config/config.json.example")
	conf.Check()

	server, ok := conf.Servers["freenode"]
	if !ok {
		t.Fatal("Can't find specified server")
	}
	irc = New(server, &wg)

	if irc.Server == nil {
		t.Fatalf("Server is nil: %+v\n", irc)
	}

	if len(irc.Events) == 0 {
		t.Errorf("No events added: %+v\n", irc)
	}
}

func TestIRC_Connect(t *testing.T) {
	if err := irc.Connect(); err != nil {
		t.Fatal("Couldn't connect to server", err)
	}

	if irc.conn == nil {
		t.Fatalf("Connection is nil: %+v\n", irc)
	}
}

func TestIRC_AddEventCallback(t *testing.T) {
	noticesLen := len(irc.Events["NOTICE"])
	irc.AddEventCallback("NOTICE", func(event *Event) {
		t.Log("Received a notice from", event.Source)
	})

	if len(irc.Events["NOTICE"]) != noticesLen+1 {
		t.Error("Did not add an event callback")
	}
}

func TestIRC_Loop(t *testing.T) {
	go irc.Loop()
}

func TestIRC_Identify(t *testing.T) {
	irc.Identify()
	irc.Server.NickservPassword = "test"
	irc.Identify()
	irc.Server.NickservPassword = ""
}

func TestIRC_Join(t *testing.T) {
	var w sync.WaitGroup

	w.Add(1)
	irc.AddEventCallback("376", func(event *Event) {
		irc.Join(channel)
		w.Done()
	})

	w.Add(1)
	irc.AddEventCallback("JOIN", func(event *Event) { // 366
		w.Done()
	})
	w.Wait()
}

func TestIRC_PrivMsgf(t *testing.T) {
	var w sync.WaitGroup

	w.Add(1)
	irc.AddEventCallback("PRIVMSG", func(event *Event) {
		w.Done()
	})

	irc.PrivMsgf(channel, "sending test message to %q", channel)
	irc.PrivMsgf(irc.Server.Nickname, "sending test message to %q", channel)

	w.Wait()
}

func TestIRC_Part(t *testing.T) {
	var w sync.WaitGroup

	irc.AddEventCallback("442", func(event *Event) {
		t.Error("Parting channel we were not into")
	})

	w.Add(1)
	irc.AddEventCallback("PART", func(event *Event) {
		w.Done()
	})

	irc.Part(channel)

	w.Wait()
}

func TestIRC_Quit(t *testing.T) {
	irc.Quit()
	wg.Wait()

	if irc.State.Registered == false {
		t.Error("Bot did not register")
	}

	if irc.State.Disconnected.Value == false {
		t.Error("Disconnected state is false, should be true")
	}

	if irc.State.Disconnected.Value == true && irc.State.Disconnected.Requested == false {
		t.Error("Bot quit unexpectedly")
	}
}

func TestIRC_Disconnect(t *testing.T) {
	err := irc.Connect()
	if err != nil {
		t.Fatal("Error (re)connecting to server", err)
	}

	if irc.State.Disconnected.Value == true {
		t.Error("Disconnected state should not be true")
	}

	if irc.State.Disconnected.Requested == true {
		t.Error("Disconnected should not be requested")
	}

	go irc.Loop()
	time.Sleep(3 * time.Second)

	irc.Disconnect("")

	if irc.State.Disconnected.Value == false {
		t.Error("Disconnected state should not be false")
	}

	if irc.State.Disconnected.Requested == false {
		t.Error("Disconnected should be requested")
	}
}
