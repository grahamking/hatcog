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
    term.Write([]byte("Terminal set to raw, listening for keyboard input\n\r"))

    // Connect to go-connect
    var conn *InternalConnection
    conn = NewInternalConnection(GO_HOST, *channel)
    defer conn.Close()

    term.Write([]byte("Connected to go-connect\n\r"))
	go conn.Consume()

    var line *Line
    for isRunning {

        select {
            case serverData := <-fromServer:
                line = FromJson(serverData)
                line.Display(term)

            case userInput := <-fromUser:
                doInput(userInput, conn)

                // Display locally
                line := Line{
                    User:"me",
                    Content:string(userInput),
                    Channel:conn.Channel}
                line.Display(term)
        }
    }

    // Internal listener for user input from socket
	//go listenInternalSocket()
	//fmt.Println("Use 'netcat 127.0.0.1 " + INTERNAL_PORT + "' to connect for writes")
}

// Act on user input
func doInput(content []byte, conn *InternalConnection) {

    conn.Write([]byte(conn.Channel + " "))
    conn.Write(content)

    if string(content) == "/quit" {
        isRunning = false
        return
    }
}

