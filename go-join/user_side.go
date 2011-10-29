package main

import (
	"net"
	"os"
)

func listenInternal(ircConn *Connection) {
    var data = make([]byte, 256)
	var content string

    for {
        data = <-inputChannel

        content = sane(string(data))

        if isCommand(content) {

            if content == "/quit" {
                ircConn.Quit()
                return
            }

            ircConn.doCommand(content)

        } else {
            ircConn.doMsg(content)
        }
    }
}

// Is 'content' an IRC command
func isCommand(content string) bool {
	return len(content) > 1 && content[0] == '/'
}


// Act as a server, forward data to irc connection
func listenInternalSocket() {
	var listener net.Listener
	var internalConn net.Conn
	var data []byte
	var err os.Error

	listener, err = net.Listen("tcp", "127.0.0.1:"+INTERNAL_PORT)

	if err != nil {
		panic("Error on internal listen:" + err.String())
	}
	defer listener.Close()

	for {
		internalConn, err = listener.Accept()
		if err != nil {
			panic("Listener accept error:" + err.String())
			break
		}
		for {

			data = make([]byte, 256)
			_, err = internalConn.Read(data)
			if err != nil {
				panic("Interal conn consume error:" + err.String())
				break
			}

            inputChannel <- data
		}
	}

}
