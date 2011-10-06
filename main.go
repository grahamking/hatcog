package main

import (
	"flag"
    "fmt"
)

const (
	INTERNAL_PORT      = "8790"
	FULL_NAME          = "Go IRC"
    IRC_NAME_LENGTH    = 15
)

var server = flag.String("server", "127.0.0.1:6667", "IP address or hostname and optional port for IRC server to connect to")
var nick = flag.String("nick", "goirc", "Nick name")
var channel = flag.String("channel", "#test", "Channel to connect to");


/*
 * main
 */
func main() {

	flag.Parse()

    // Set terminal to raw mode, listen for keyboard input
    var term *Terminal = NewTerminal()
    defer func() {
        term.Close()
        fmt.Println("Bye!")
    }()
    term.Raw()

    // IRC connection to remote server
    var conn *Connection
    conn = NewConnection(*server, *nick, FULL_NAME, *channel, term)
	defer conn.Close()

    term.Channel = *channel

    // Gathers all inputs and sends to IRC server
	go listenInternal(conn)

    // Internal listener for user input from socket
	go listenInternalSocket()
	fmt.Println("Use 'netcat 127.0.0.1 " + INTERNAL_PORT + "' to connect for writes")

    // Internal listener for user input from keyboard
    go term.ListenInternalKeys()

    // External (IRC server) consume
	conn.Consume()
}

