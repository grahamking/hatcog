package main

import (
    "net"
    "os"
    "log"
)

const (
    ONE_SECOND_NS = 1000 * 1000 * 1000
)

type InternalConnection struct {
    socket net.Conn
    Channel string
}

func NewInternalConnection(host string, channel string) *InternalConnection {

    var socket net.Conn
    var err os.Error
    socket, err = net.Dial("tcp", host)
    if err != nil {
        log.Fatal("Error connection to go-connect", err)
    }

    socket.SetReadTimeout(ONE_SECOND_NS)

    conn := InternalConnection{socket, channel}
    conn.Write([]byte("/join " + channel))

    return &conn
}

// Send a message to go-connect. Implements Writer.
func (self *InternalConnection) Write(msg []byte) (int, os.Error) {
    bytesWritten, err := self.socket.Write([]byte(msg))
    return bytesWritten, err
}

// Listen for JSON messages from go-connect and put on channel
func (self *InternalConnection) Consume() {
	var data []byte = make([]byte, 1)
	var linedata []byte = make([]byte, 4096)
    var err os.Error
    var index int

    for {

		_, err = self.socket.Read(data)
		if err != nil {
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
