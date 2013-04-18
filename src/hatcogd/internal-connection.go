package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net"
	"strings"
)

type Internal struct {
	netConn   net.Conn
	channel   string // channel or nick (for private/query messages)
	network   string // address of remote we use (e.g. "irc.freenode.net:6697")
	isPrivate bool   // if True, channel is the nick
	manager   *InternalManager
}

func (self *Internal) Run() {
	defer logPanic()

	// Send NICK msg to new client connections
	self.sendNick()

	for {

		bufRead := bufio.NewReader(self.netConn)
		content, err := bufRead.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Leaving", self.channel)
				self.part()
				self.manager.delete(self)
			} else {
				log.Println(err)
			}
			return
		}
		content = content[:len(content)-1] // Chop \n

		if self.Special(content) {
			continue
		}

		self.manager.fromUser <- Message{self.network, self.channel, content}
	}
}

/* Special incoming command processing, used to implement
non-standard function, mostly about communication between hjoin
and client.

@return true if no further processing should occur, false otherwise. true
means don't send this message to the IRC server, it was internal only.

TODO: This should move to Server.onUser.
*/
func (self *Internal) Special(content string) bool {

	var parts []string

	if self.network == "" && strings.HasPrefix(content, "/connect") {
		parts = strings.Split(content, " ")
		if len(parts) == 2 {
			log.Println("parts1: ", parts[1])
			self.network, _ = splitNetPass(parts[1])
			log.Println("Network is", self.network)
		}
	}

	if self.channel == "" {

		isPrivate := strings.HasPrefix(content, "/private")
		self.isPrivate = isPrivate

		isJoin := strings.HasPrefix(content, "/join")

		if isJoin || isPrivate {
			parts = strings.Split(content, " ")
			if len(parts) == 2 {
				self.channel = parts[1]
			}
		}

		if isPrivate {
			// Send most recent private message, so new window shows it
			if self.manager.lastPrivate != nil {
				self.netConn.Write(self.manager.lastPrivate)
				self.manager.lastPrivate = nil
			}

			// Not an IRC server command, so don't send to IRC server
			return true
		}
	}

	return false
}

func (self *Internal) sendNick() {

	nick := self.manager.GetNick(self.network)
	if nick == "" {
		return
	}

	line := Line{
		Command: "NICK",
		Content: nick,
		Channel: "",
		User:    ""}

	jsonData, _ := json.Marshal(line)
	self.netConn.Write(append(jsonData, '\n'))
}

// Client closed connection, leave channel, no more work here
func (self *Internal) part() {

	if self.channel == "" || self.isPrivate {
		return
	}

	self.manager.fromUser <- Message{self.network, self.channel, "/part " + self.channel}
}
