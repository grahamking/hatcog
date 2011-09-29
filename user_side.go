package main

import (
	"net"
	"fmt"
	"os"
)

// Act as a server, forward data to irc connection
func listenInternal(ircConn *Connection) {
	var listener net.Listener
	var internalConn net.Conn
	var data []byte
	var err os.Error
	var content string

	listener, err = net.Listen("tcp", "127.0.0.1:"+INTERNAL_PORT)
	fmt.Println("Use 'netcat 127.0.0.1 " + INTERNAL_PORT + "' to connect for writes")

	if err != nil {
		println("Error on internal listen:", err)
	}
	defer listener.Close()

	for {
		internalConn, err = listener.Accept()
		if err != nil {
			println("Listener accept error:", err)
			break
		}
		for {

			data = make([]byte, 1024)
			_, err = internalConn.Read(data)
			if err != nil {
				println("Interal conn consume error:", err)
				break
			}
			content = sane(string(data))

			if isCommand(content) {
				ircConn.doCommand(content)
			} else {
				ircConn.doMsg(content)
			}

		}
	}

}

// Is 'content' an IRC command
func isCommand(content string) bool {
	return len(content) > 1 && content[0] == '/'
}
