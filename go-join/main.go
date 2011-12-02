package main

import (
	"fmt"
	"os"
	"log"
	"strings"
	"flag"
)

const (
	// go-connect and go-join must be on same host for now,
	/// but in future go-connect could be remote
	GO_HOST            = "127.0.0.1:8790"
	RPL_NAMREPLY       = "353"
	RPL_TOPIC          = "332"
	ERR_UNKNOWNCOMMAND = "421"
	CHANNEL_CMDS       = "PRIVMSG, ACTION, PART, JOIN, " + RPL_NAMREPLY
	//CMD_PRIV_CHAT = "/usr/bin/tmux split-window -v -p 50"
	CMD_PRIV_CHAT = "/usr/bin/gnome-terminal -e"
	USAGE         = `
Usage: go-join [channel|-private=nick]
Note there's no # in front of the channel
Examples:
 1. Join channel test: go-join test
 2. Listen for private (/query) message from bob: go-join -private=bob
`
)

var userPrivate = flag.String(
	"private",
	"",
	"Listen for private messages from this nick only")

var fromUser = make(chan []byte)
var fromServer = make(chan []byte)

// Logs messages from go-connect
var rawLog *log.Logger

func init() {
	var logfile *os.File
	logfile, _ = os.Create("/tmp/go-join.log")
	rawLog = log.New(logfile, "", log.LstdFlags)
}

/*
 * main
 */
func main() {

	if len(os.Args) != 2 {
		fmt.Println(USAGE)
		os.Exit(1)
	}

	var channel string

	arg := os.Args[1]
	if strings.HasPrefix(arg, "-private") {
		flag.Parse()
		channel = *userPrivate
	} else {
		channel = "#" + arg
	}

	client := NewClient(channel)
	defer func() {
		client.Close()
		fmt.Println("Bye!")
	}()

	client.Run()
}

