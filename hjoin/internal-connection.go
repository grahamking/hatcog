package main

import (
	"net"
	"os"
	"log"
	"strings"
)

const (
	ONE_SECOND_NS = 1000 * 1000 * 1000
)

type InternalConnection struct {
	socket  net.Conn
	channel string
}

func NewInternalConnection(host string, channel string) *InternalConnection {

    socket, err := net.Dial("tcp", host)
	if err != nil {
        // TODO: Start the daemon!
		LOG.Fatal("Error connecting to daemon", err)
	}

	socket.SetReadTimeout(ONE_SECOND_NS)

	return &InternalConnection{socket, channel}
}

// Join a channel, or tell daemon we're a private chat
func (self *InternalConnection) join() {

	if strings.HasPrefix(self.channel, "#") {
		// Join a channel
		self.Write([]byte("/join " + self.channel))

	} else {
		// Private message (query). /private is not standard.
		self.Write([]byte("/private " + self.channel))
	}
}

// Send a message to daemon. Implements Writer.
func (self *InternalConnection) Write(msg []byte) (int, os.Error) {
	return self.socket.Write(append(msg, '\n'))
}

// Listen for JSON messages from daemon and put on channel
func (self *InternalConnection) Consume() {
	var data []byte = make([]byte, 1)
	var linedata []byte = make([]byte, 4096)
	var err os.Error
	var index int

	self.join()

	for {

		_, err = self.socket.Read(data)
		if err != nil {
			if err == os.EOF {
				// Internal connection closed
				close(fromServer)
				return
			}

			netErr, _ := err.(net.Error)

			// Need to timeout occasionally or we never check isClosing
			if netErr.Timeout() == true {
				continue
			} else {
				log.Fatal("Consume Error:", err)
			}
		}

		if data[0] == '\n' {
			fromServer <- linedata[:index]
			index = 0
		} else if data[0] != '\r' { // Ignore CR, because LF is next
			linedata[index] = data[0]
			index++
		}
	}
}

// Close connection to go-connect
func (self *InternalConnection) Close() os.Error {
	return self.socket.Close()
}
