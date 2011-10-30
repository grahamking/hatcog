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

// IRC Channel command line param
var channel = flag.String("channel", "#test", "Channel to connect to");

var isRunning = true

var fromUser = make(chan []byte)
var fromServer = make(chan []byte)

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
    term.Channel = *channel
    go term.ListenInternalKeys()
    fmt.Println("Terminal set to raw, listening for keyboard input")

    // Connect to go-connect
    var conn *InternalConnection
    conn = NewInternalConnection(GO_HOST, *channel)
    defer conn.Close()

    fmt.Println("Connected to go-connect")
	go conn.Consume()

    for isRunning {

        select {
            case serverData := <-fromServer:
                // TODO: serverData is JSON, decode it and display
                term.Write(append(serverData, '\n'))

            case userInput := <-fromUser:
                doInput(userInput, conn)
        }
    }

    // Internal listener for user input from socket
	//go listenInternalSocket()
	//fmt.Println("Use 'netcat 127.0.0.1 " + INTERNAL_PORT + "' to connect for writes")

    fmt.Println("Bye")
}

// Act on user input
func doInput(content []byte, conn *InternalConnection) {

    conn.Write([]byte("#" + conn.Channel + " "))
    conn.Write(content)

    if string(content) == "/quit\n" {
        isRunning = false
        return
    }
}

