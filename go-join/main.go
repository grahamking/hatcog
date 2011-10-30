package main

import (
	"flag"
    "fmt"
    "os"
)

const (
    // go-connect and go-join must be on same host for now,
    /// but in future go-connect could be remote
    GO_HOST      = "127.0.0.1:8790"
	RPL_NAMREPLY = "353"
    JOIN = "JOIN"
    NICK = "NICK"
)

// IRC Channel command line param
var channel = flag.String("channel", "#test", "Channel to connect to");

var fromUser = make(chan []byte)
var fromServer = make(chan []byte)

/*
 * main
 */
func main() {

    flag.Parse()

    client := NewClient(*channel)
    defer func() {
        client.Close()
        fmt.Println("Bye!")
    }()

    client.Run()
}

// IRC Client abstraction
type Client struct {
    term *Terminal
    conn *InternalConnection
    channel string
    isRunning bool
    nick string
}

// Create IRC client. Switch keyboard to raw mode, connect to go-connect socket
func NewClient(channel string) *Client {

    // Set terminal to raw mode, listen for keyboard input
    var term *Terminal = NewTerminal()
    term.Raw()
    term.Channel = channel

    // Connect to go-connect
    var conn *InternalConnection
    conn = NewInternalConnection(GO_HOST, channel)

    return &Client{term: term, conn: conn, channel: channel, nick: "goirc"}
}

/* Main loop
   Listen for keyboard input and socket input and be an IRC client
*/
func (self *Client) Run() {

    self.isRunning = true

    go self.term.ListenInternalKeys()
    self.display("Listening for keyboard input")

	go self.conn.Consume()
    self.display("Connected to go-connect")
    self.display("Joining channel " + self.channel)

    for self.isRunning {

        select {
            case serverData := <-fromServer:
                self.onServer(serverData)

            case userInput := <-fromUser:
                self.onUser(userInput)
        }
    }

    // Internal listener for user input from socket
	//go listenInternalSocket()
	//fmt.Println("Use 'netcat 127.0.0.1 " + INTERNAL_PORT + "' to connect for writes")

    return
}

// Do something with user input. Usually just send to go-connect
func (self *Client) onUser(content []byte) {

    if string(content) == "/quit" {
        // Local quit, don't send to server
        // Currently there's no global quit
        self.isRunning = false
        return
    }

    if isCommand(content) {
        // IRC command
        self.conn.Write(content)
    } else {
        // Send to go-connect
        self.conn.Write([]byte(self.channel + " "))
        self.conn.Write(content)

        // Display locally
        line := Line{
            User:self.nick,
            Content:string(content),
            Channel:self.channel}
        self.term.Write([]uint8(line.String()))
    }

}

// Do something with Line from go-connect. Usually just display to screen.
func (self *Client) onServer(serverData []byte) {

    line := FromJson(serverData)
    if line.Channel != self.channel {
        return
    }

    if line.HasDisplay() {
        self.term.Write([]byte(line.String()))
    }

    // TODO
    // - JOIN is only sent the first time
    // we connect to a channel. If we hop back
    // it doesn't get sent because unless we /part,
    // we are still connected. So can't rely on it
    // to update internal state, instead
    // need to parse the command ourselves.
    // Or maybe server sends something else?
    // - Record messages even if not being displayed right now.
    // - JOIN is sent for every other user that connects too.
    if line.Command == JOIN {
        if line.User == self.nick {
            if len(line.Content) != 0 {
                self.channel = line.Content
            } else if len(line.Args) != 0 {
                self.channel = line.Args[0]
            }
            self.term.Channel = self.channel
            self.display("Now talking on " + self.channel)
        } else {
            self.display("User joined channel: " + line.User)
        }

    } else if line.Command == RPL_NAMREPLY {
        self.display("Users currently in " + self.channel + ": ")
        self.display(line.Content)

    } else if line.Command == NICK {
        if line.User == self.nick {
            self.nick = line.Content
            self.display("You are now know as" + self.nick)
        } else {
            self.display(line.User + "is now know as" + line.Content)
        }
    }
}

// Write string to terminal
func (self *Client) display(msg string) {
    self.term.Write([]byte(msg + "\n\r"))
}

func (self *Client) Close() os.Error {
    self.term.Close()
    return self.conn.Close()
}

// Is 'content' an IRC command?
func isCommand(content []byte) bool {
	return len(content) > 1 && content[0] == '/'
}
