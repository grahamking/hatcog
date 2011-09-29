package main

import (
	"net"
	"os"
	"log"
	"time"
	"fmt"
)

const (
    ONE_SECOND_NS = 1000 * 1000 * 1000
	RPL_NAMREPLY = "353"
    JOIN = "JOIN"
    PING = "PING"
    NICK = "NICK"
)

type Connection struct {
	socket  net.Conn
	channel string
	nick    string
	name    string
}

func NewConnection(server string, nick string, name string, channel string) *Connection {

	var socket net.Conn
	var err os.Error
	socket, err = net.Dial("tcp", server)
	if err != nil {
		log.Fatal("Error on IRC connect:", err)
	}
	time.Sleep(ONE_SECOND_NS)

	conn := Connection{socket, channel, nick, name}
	conn.SendRaw("USER " + nick + " localhost localhost :" + name)
	conn.SendRaw("NICK " + nick)
	time.Sleep(ONE_SECOND_NS)

	conn.SendRaw("JOIN " + channel)

	return &conn
}

// Send a regular (non-system command) IRC message
func (self *Connection) SendMessage(msg string) {
	fullmsg := "PRIVMSG " + self.channel + " :" + msg
	self.SendRaw(fullmsg)
}

// Send message down socket
func (self *Connection) SendRaw(msg string) {
	var full = msg + "\n"
	var err os.Error

    rawLog.Println(" --> ", msg);

    if *isRaw {
        log.Print(full)
    }

	_, err = self.socket.Write([]byte(full))
	if err != nil {
		log.Fatal("Error writing to socket", err)
	}
}

// Process a slash command
func (self *Connection) doCommand(content string) {

	content = content[1:]
	self.SendRaw(content)
	/*
    var parts []string
    var command, rest string

    parts = strings.SplitN(content[1:], " ", 2)
    command = parts[0]
    rest = parts[1]

    if strings.ToUpper(command) == "JOIN" {
        doJoin(ircConn, rest)
    }
	*/
}

// Process a user message
func (self *Connection) doMsg(content string) {
	// Send to server
	self.SendMessage(content)

	// Display for ourselves
	msg := "< " + self.nick + "> " + content
	fmt.Println(msg)
}

// Read IRC messages from the connection and send to stdout
func (self *Connection) Consume() {
	var data []byte = make([]byte, 1)
	var linedata []byte = make([]byte, 4096)
	var index int
	var line Line
	var err os.Error = nil

	for {
		_, err = self.socket.Read(data)
		if err != nil {
			log.Fatal("Consume Error:", err)
		}

		if data[0] == '\n' {
			line = ParseLine(string(linedata[:index]))
            self.act(line)

			index = 0
		} else if data[0] != '\r' { // Ignore CR, because LF is next
			linedata[index] = data[0]
			index++
		}
	}
}

// Do something with a line
func (self *Connection) act(line Line) {

    if line.HasDisplay() {
        fmt.Println(line.String())
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
            fmt.Println("Now talking on", self.channel)
        } else {
            fmt.Println("User joined channel:", line.User)
        }

    } else if line.Command == PING {
        self.SendRaw("PONG goirc");

    } else if line.Command == RPL_NAMREPLY {
        fmt.Print("Users currently in ", self.channel, ": ")
        fmt.Println(line.String())

    } else if line.Command == NICK {
        if line.User == self.nick {
            self.nick = line.Content
            fmt.Println("You are now know as", self.nick)
        } else {
            fmt.Println(line.User, "is now know as", line.Content)
        }
    }
}

func (self *Connection) Close() {
	self.socket.Close()
}
