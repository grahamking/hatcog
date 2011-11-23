package main

import (
	"net"
	"os"
	"json"
	"strings"
	"fmt"
    "bufio"
)

type Message struct {
    channel string
    content string
}

type Internal struct {
	port        string
	connections []*InternalConnection
	fromUser    chan Message
	Nick        string // Need to know, to tell go-join
}

type InternalConnection struct {
	netConn   net.Conn
	channel   string // channel or nick (for private/query messages)
	isPrivate bool   // if True, channel is the nick
}

func NewInternal(port string, fromUser chan Message, nick string) *Internal {

	var connections = make([]*InternalConnection, 0)
	return &Internal{port, connections, fromUser, nick}
}

// Act as a server, forward data to irc connection
func (self *Internal) Run() {

	var listener net.Listener
	var netConn net.Conn
	var internalConn *InternalConnection
	var err os.Error

	listener, err = net.Listen("tcp", "127.0.0.1:"+self.port)

	if err != nil {
		panic("Error on internal listen: " + err.String())
	}
	defer listener.Close()

	for {
		netConn, err = listener.Accept()
		if err != nil {
			panic("Listener accept error: " + err.String())
			break
		}

		internalConn = &InternalConnection{netConn: netConn, channel: ""}
		self.connections = append(self.connections, internalConn)
		go self.workConnection(internalConn)
	}

}

func (self *Internal) workConnection(internalConn *InternalConnection) {

	// Send NICK msg to new go-join connections
	self.sendNick(internalConn.netConn)

	for {

		bufRead := bufio.NewReader(internalConn.netConn)
        content, err := bufRead.ReadString('\n')
		if err != nil {
			if err == os.EOF {
				self.part(internalConn)
				self.delete(internalConn)
			} else {
				fmt.Println(err)
			}
			return
		}
        content = content[:len(content)-1]      // Chop \n
        fmt.Println("Content in: " + content)

        // First message from go-join is either /join or /private,
        // telling us which channel or user this go-join is talking to
        if internalConn.channel == "" {

            isPrivate := strings.HasPrefix(content, "/private")
            internalConn.isPrivate = isPrivate

            isJoin := strings.HasPrefix(content, "/join")

            if isJoin || isPrivate {
                fmt.Println("isPrivate or isJoin")
                parts := strings.Split(content, " ")
                if len(parts) == 2 {
                    internalConn.channel = parts[1]
                }
            }
            if isPrivate {  // Not an IRC server command
                continue
            }
        }

		self.fromUser <- Message{internalConn.channel, content}
	}
}

func (self *Internal) sendNick(netConn net.Conn) {

	line := Line{Command: "NICK", Content: self.Nick, Channel: "", User: ""}

	jsonData, _ := json.Marshal(line)
	netConn.Write(append(jsonData, '\n'))
}

// Write a message to channel connection
func (self *Internal) WriteChannel(channel string, msg []byte) (int, os.Error) {

    var bytesWritten int

	for _, conn := range self.connections {
		if conn.channel == channel {
            conn.netConn.Write(msg)
            bytesWritten += len(msg)
        }
	}

    if bytesWritten == 0 && len(self.connections) > 0 {
        // No matching channel yet, send to first connection
        self.connections[0].netConn.Write(msg)
        bytesWritten += len(msg)
    }

	return bytesWritten, nil
}

// Write a message to all go-join connections
func (self *Internal) WriteAll(msg []byte) (int, os.Error) {

    var bytesWritten int
	for _, conn := range self.connections {
        conn.netConn.Write(msg)
        bytesWritten += len(msg)
	}

    return bytesWritten, nil
}

// Client closed connection, leave channel, no more work here
func (self *Internal) part(internalConn *InternalConnection) {

	if internalConn.channel == "" || internalConn.isPrivate {
		return
	}

	self.fromUser <- Message{internalConn.channel, "/part " + internalConn.channel}
}

// Remove a connection from our list, probably because user closed it
func (self *Internal) delete(internalConn *InternalConnection) {
	newConnections := make([]*InternalConnection, 0, len(self.connections) - 1)
	for _, conn := range self.connections {
		if conn.channel != internalConn.channel {
			newConnections = append(newConnections, conn)
		}
	}
	self.connections = newConnections
}

func (self *Internal) Close() os.Error {
	for _, conn := range self.connections {
		conn.netConn.Close()
	}
	return nil
}
