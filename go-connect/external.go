package main

import (
	"net"
	"os"
	"log"
	"time"
)

const (
    ONE_SECOND_NS = 1000 * 1000 * 1000
	RPL_NAMREPLY = "353"
    PING = "PING"
)

type lineCallback func(line *Line)

type Connection struct {
	socket  net.Conn
	name    string
    isClosing bool
    fromServer chan *Line
}


func NewConnection(server string, nick string, name string, fromServer chan *Line) *Connection {

	var socket net.Conn
	var err os.Error
	socket, err = net.Dial("tcp", server)
	if err != nil {
		log.Fatal("Error on IRC connect:", err)
	}
	time.Sleep(ONE_SECOND_NS)

    socket.SetReadTimeout(ONE_SECOND_NS)

	conn := Connection{
        socket: socket,
        name: name,
        fromServer: fromServer,
    }
	conn.SendRaw("USER " + nick + " localhost localhost :" + name)
	conn.SendRaw("NICK " + nick)
	time.Sleep(ONE_SECOND_NS)

	return &conn
}

// Send a regular (non-system command) IRC message
func (self *Connection) SendMessage(channel, msg string) {
	fullmsg := "PRIVMSG " + channel + " :" + msg
	self.SendRaw(fullmsg)
}

// Send message down socket
func (self *Connection) SendRaw(msg string) {
	var full = msg + "\n"
	var err os.Error

    rawLog.Println(" -->", msg);

	_, err = self.socket.Write([]byte(full))
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
func (self *Connection) act(line *Line) {

    if line.Command == PING {
        self.SendRaw("PONG goirc");
        return
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
