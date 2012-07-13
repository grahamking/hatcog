package main

import (
	"log"
	"net"
)

type InternalManager struct {
	host        string
	port        string
	connections []*Internal
	fromUser    chan Message
	Nick        string // Need to know, to tell client
	lastPrivate []byte // Most recent private message
}

type Message struct {
	network string
	channel string
	content string
}

func NewInternalManager(host, port string, fromUser chan Message) *InternalManager {

	var connections = make([]*Internal, 0)
	return &InternalManager{host, port, connections, fromUser, "", nil}
}

// Act as a server, forward data to irc connection
func (self *InternalManager) Run() {

	var listener net.Listener
	var netConn net.Conn
	var internalConn *Internal
	var err error

	listener, err = net.Listen("tcp", self.host+":"+self.port)

	if err != nil {
		log.Fatal("Error on internal listen: " + err.Error())
	}
	defer listener.Close()

	for {
		netConn, err = listener.Accept()
		if err != nil {
			log.Fatal("Listener accept error: " + err.Error())
			break
		}

		internalConn = &Internal{
			netConn: netConn,
			channel: "",
			network: "",
			manager: self}
		self.connections = append(self.connections, internalConn)
		go internalConn.Run()
	}

}

// Write a message to channel connection
func (self *InternalManager) WriteChannel(network, channel string, msg []byte) (int, error) {

	var bytesWritten int

	for _, conn := range self.connections {
		if conn.channel == channel && conn.network == network {
			conn.netConn.Write(msg)
			bytesWritten += len(msg)
		}
	}

	return bytesWritten, nil
}

// Write a message to all client connections on a given network
func (self *InternalManager) WriteAll(network string, msg []byte) (int, error) {

	var bytesWritten int
	for _, conn := range self.connections {
		if conn.network == network {
			conn.netConn.Write(msg)
			bytesWritten += len(msg)
		}
	}

	return bytesWritten, nil
}

// Write a message to only the first channel.
// This is used when a message has to go to the client, but it doesn't
// matter which one it goes to. Used to open a private chat window.
func (self *InternalManager) WriteFirst(network string, msg []byte) (int, error) {

    var bytesWritten int
    var err error

	for _, conn := range self.connections {
		if conn.network == network {
            bytesWritten, err = conn.netConn.Write(msg)
            break
		}
	}

	return bytesWritten, err
}

// Remove a connection from our list, probably because user closed it
func (self *InternalManager) delete(internalConn *Internal) {
	if len(self.connections) == 0 {
		return
	}

	newConnections := make([]*Internal, 0, len(self.connections)-1)
	for _, conn := range self.connections {
		if conn.channel != internalConn.channel {
			newConnections = append(newConnections, conn)
		}
	}
	self.connections = newConnections
}

// The internal connection for given channel, or nil
func (self *InternalManager) GetChannelConnection(channel string) *Internal {
	for _, conn := range self.connections {
		if conn.channel == channel {
			return conn
		}
	}
	return nil
}

// Do we have a connection (a client) open on given channel or nick
func (self *InternalManager) HasChannel(channel string) bool {
	return self.GetChannelConnection(channel) != nil
}

func (self *InternalManager) Close() error {
	for _, conn := range self.connections {
		conn.netConn.Close()
	}
	return nil
}
