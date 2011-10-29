package main

import (
	"flag"
    "fmt"
)

const (
    // go-connect and go-join must be on same host for now,
    /// but in future go-connect could be remote
    GO_HOST      = "127.0.0.1:8790"

    IRC_NAME_LENGTH    = 15
)

// IRC Channel
var channel = flag.String("channel", "#test", "Channel to connect to");

// Go comms channel
var inputChannel = make(chan []byte)

/*
 * main
 */
func main() {

    // Set terminal to raw mode, listen for keyboard input
    var term *Terminal = NewTerminal()
    defer func() {
        term.Close()
        fmt.Println("Bye!")
    }()
    term.Raw()

    // Connect to go-connect

    var conn * InternalConnection
    conn = NewInternalConnection(GO_HOST, *channel, term)
    defer conn.Close()

    term.Channel = *channel

    // Gathers all inputs and sends to IRC server
	//go listenInternal(conn)

    // Internal listener for user input from socket
	//go listenInternalSocket()
	//fmt.Println("Use 'netcat 127.0.0.1 " + INTERNAL_PORT + "' to connect for writes")

    // Internal listener for user input from keyboard
    //go term.ListenInternalKeys()

    // Listen for messages from go-connect
	conn.Receive()
}

