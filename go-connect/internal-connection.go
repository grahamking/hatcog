package main

import (
	"net"
	"os"
	"json"
	"strings"
	"fmt"
    "bufio"
)

type Internal struct {
	netConn   net.Conn
	channel   string    // channel or nick (for private/query messages)
	isPrivate bool      // if True, channel is the nick
    manager   *InternalManager
}

func (self *Internal) Run() {

	// Send NICK msg to new go-join connections
	self.sendNick()

	for {

		bufRead := bufio.NewReader(self.netConn)
        content, err := bufRead.ReadString('\n')
		if err != nil {
			if err == os.EOF {
				self.part()
                self.manager.delete(self)   // TODO: Replace with channel?
			} else {
				fmt.Println(err)
			}
			return
		}
        content = content[:len(content)-1]      // Chop \n

        // First message from go-join is either /join or /private,
        // telling us which channel or user this go-join is talking to
        if self.channel == "" {

            isPrivate := strings.HasPrefix(content, "/private")
            self.isPrivate = isPrivate

            isJoin := strings.HasPrefix(content, "/join")

            if isJoin || isPrivate {
                parts := strings.Split(content, " ")
                if len(parts) == 2 {
                    self.channel = parts[1]
                }
            }
            if isPrivate {  // Not an IRC server command
                continue
            }
        }

		self.manager.fromUser <- Message{self.channel, content}
	}
}

func (self *Internal) sendNick() {

	line := Line{
        Command: "NICK",
        Content: self.manager.Nick,
        Channel: "",
        User: ""}

	jsonData, _ := json.Marshal(line)
	self.netConn.Write(append(jsonData, '\n'))
}

// Client closed connection, leave channel, no more work here
func (self *Internal) part() {

	if self.channel == "" || self.isPrivate {
		return
	}

	self.manager.fromUser <- Message{self.channel, "/part " + self.channel}
}

