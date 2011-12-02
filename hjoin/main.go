package main

import (
	"fmt"
	"os"
	"log"
	"strings"
	"flag"
)

const (
	VERSION    = "hatcog v0.3 (github.com/grahamking/hatcog)"
    DEFAULT_CONFIG = "/.hatcogrc"
    LOG_DIR = "/.hatcog/"

	GO_HOST            = "127.0.0.1:8790"

	RPL_NAMREPLY       = "353"
	RPL_TOPIC          = "332"
	ERR_UNKNOWNCOMMAND = "421"
	CHANNEL_CMDS       = "PRIVMSG, ACTION, PART, JOIN, " + RPL_NAMREPLY

	USAGE         = `
Usage: hjoin [channel|-private=nick]
Note there's no # in front of the channel
Examples:
 1. Join channel test: go-join test
 2. Listen for private (/query) message from bob: go-join -private=bob
`
)

var (
    HOME string
    LOG *log.Logger
)

var userPrivate = flag.String(
	"private",
	"",
	"Listen for private messages from this nick only")

var fromUser = make(chan []byte)
var fromServer = make(chan []byte)

/*
 * main
 */
func main() {

    HOME = os.Getenv("HOME")

    logFilename := HOME + LOG_DIR + "client.log"
    fmt.Println(VERSION, "logging to", logFilename)
    LOG = openLog(logFilename)

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

// Open the main log file
func openLog(logFilename string) *log.Logger {
    os.Mkdir(HOME + LOG_DIR, 0750)

    logFile, err := os.OpenFile(
        logFilename,
        os.O_RDWR|os.O_APPEND|os.O_CREATE,
        0650)
    if err != nil {
        fmt.Println("Error creating log file:", logFilename, err)
        os.Exit(1)
    }
    return log.New(logFile, "", log.LstdFlags)
}

