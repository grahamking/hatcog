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
	RPL_NAMREPLY  = "353"
	PING          = "PING"
)

type Connection struct {
	socket     net.Conn
	name       string
	isClosing  bool
	fromServer chan *Line
}

func NewConnection(server string,
nick string,
name string,
password string,
fromServer chan *Line) *Connection {

	var socket net.Conn
	var err os.Error
	socket, err = net.Dial("tcp", server)
	if err != nil {
		log.Fatal("Error on IRC connect:", err)
	}
	time.Sleep(ONE_SECOND_NS)

	socket.SetReadTimeout(ONE_SECOND_NS)

	conn := Connection{
		socket:     socket,
		name:       name,
		fromServer: fromServer,
	}
	conn.SendRaw("USER " + nick + " localhost localhost :" + name)
	conn.SendRaw("NICK " + nick)
	time.Sleep(ONE_SECOND_NS)

	if password != "" {
		conn.SendMessage("NickServ", "identify "+password)
	}

	return &conn
}

// Send a regular (non-system command) IRC message
func (self *Connection) SendMessage(channel, msg string) {
	fullmsg := "PRIVMSG " + channel + " :" + msg
	self.SendRaw(fullmsg)
}

// Send a /me action message
func (self *Connection) SendAction(channel, msg string) {
	fullmsg := "PRIVMSG " + channel + " :\u0001ACTION " + msg + "\u0001"
	self.SendRaw(fullmsg)
}

// Send message down socket. Add \n at end first.
func (self *Connection) SendRaw(msg string) {

	var err os.Error
	msg = msg + "\n"

	rawLog.Print(" -->", msg)

	_, err = self.socket.Write([]byte(msg))
	if err != nil {
		log.Fatal("Error writing to socket", err)
	}
}

// Process a slash command
func (self *Connection) doCommand(content string) {

	content = content[1:]
	self.SendRaw(content)
}

// Read IRC messages from the connection and send to stdout
func (self *Connection) Consume() {
	var data []byte = make([]byte, 1)
	var linedata []byte = make([]byte, 4096)
	var index int
	var line *Line
	var rawLine string
	var err os.Error = nil
	var netErr net.Error

	for {

		if self.isClosing {
			return
		}

		_, err = self.socket.Read(data)

		if err != nil {
			netErr, _ = err.(net.Error)

			// Need to timeout occasionally or we never check isClosing
			if netErr.Timeout() == true {
				continue
			} else {
				log.Fatal("Consume Error:", err)
			}
		}

		if data[0] == '\n' {
			rawLine = string(linedata[:index])
			line, err = ParseLine(rawLine)
			if err == nil {
				self.act(line)
			} else {
				rawLog.Println(err)
				fmt.Println("Invalid line: " + rawLine)
			}

			index = 0
		} else if data[0] != '\r' { // Ignore CR, because LF is next
			linedata[index] = data[0]
			index++
		}
	}
}

// Do something with a line
func (self *Connection) act(line *Line) {

	if line.Command == PING {
		self.SendRaw("PONG goirc")
		return
	} else if line.Command == "VERSION" {
		versionMsg := "NOTICE " + line.User + " :\u0001VERSION " + VERSION + "\u0001\n"
		self.SendRaw(versionMsg)
	}

	self.fromServer <- line
}

func (self *Connection) Close() os.Error {
	return self.socket.Close()
}

/* Close connection, return from event loop.
 */
func (self *Connection) Quit() {
	self.isClosing = true
}
