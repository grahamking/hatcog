package main

import (
	"net"
	"os"
	"strings"
    "bufio"
)

type InternalConnection struct {
	socket  net.Conn
	channel string
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

	self.join()

    bufRead := bufio.NewReader(self.socket)
	for {

        content, err := bufRead.ReadString('\n')

		if err != nil {
			if err == os.EOF {    // Internal connection closed
				close(fromServer)
				return
			}

			netErr, _ := err.(net.Error)

			if netErr.Timeout() == true {
				continue
			} else {
				LOG.Fatal("Consume Error:", err)
			}
		}

		content = content[:len(content)-1] // Chop \n
        fromServer <- content
	}
}

// Close connection to daemon
func (self *InternalConnection) Close() os.Error {
	return self.socket.Close()
}
