package main

import (
	"net"
	"os"
    "json"
)

type Internal struct {
    port string
    connections []net.Conn
    fromUser chan string
    Nick string             // Need to know, to tell go-join
}

func NewInternal(port string, fromUser chan string, nick string) *Internal {

    var connections = make([]net.Conn, 0)
    return &Internal{port, connections, fromUser, nick}
}

// Act as a server, forward data to irc connection
func (self *Internal) Run() {

	var listener net.Listener
	var internalConn net.Conn
	var err os.Error

    listener, err = net.Listen("tcp", "127.0.0.1:" + self.port)

	if err != nil {
		panic("Error on internal listen: " + err.String())
	}
	defer listener.Close()

	for {
		internalConn, err = listener.Accept()
		if err != nil {
			panic("Listener accept error: " + err.String())
			break
		}

        self.connections = append(self.connections, internalConn)
        go self.workConnection(internalConn)
	}

}

func (self *Internal) workConnection(internalConn net.Conn) {

    // Send NICK msg to new go-join connections
    self.sendNick(internalConn)

    for {

        data := make([]byte, 256)
        _, err := internalConn.Read(data)
        if err != nil {
            // Client closed connection, no more work here
            return
        }

        content := sane(string(data))
        self.fromUser <- content
    }
}

func (self *Internal) sendNick(internalConn net.Conn) {

    line := Line{Command: "NICK", Content: self.Nick, Channel:"" , User: ""}

    jsonData, _ := json.Marshal(line)
    internalConn.Write(append(jsonData, '\n'))
}

func (self *Internal) Write(msg []byte) (int, os.Error) {

    for _, conn := range(self.connections) {
        conn.Write(msg)
    }
    return len(msg), nil
}

func (self *Internal) Close() os.Error {
    for _, conn := range(self.connections) {
        conn.Close()
    }
	return nil
}
