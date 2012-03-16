package main

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"strconv"
	"strings"
)

type Internal struct {
	netConn   net.Conn
	channel   string // channel or nick (for private/query messages)
	isPrivate bool   // if True, channel is the nick
	isNotify  bool   // display all messages with OS notifications
	manager   *InternalManager
}

func (self *Internal) Run() {

	// Send NICK msg to new client connections
	self.sendNick()

	for {

		bufRead := bufio.NewReader(self.netConn)
		content, err := bufRead.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				LOG.Println("Leaving", self.channel)
				self.part()
				self.manager.delete(self)
			} else {
				LOG.Println(err)
			}
			return
		}
		content = content[:len(content)-1] // Chop \n

		if self.Special(content) {
			continue
		}

		self.manager.fromUser <- Message{self.channel, content}
	}
}

/* Special incoming command processing, used to implement
non-standard function, mostly about communication between go-connect
and client.

@return true if no further processing should occur, false otherwise. true
means don't send this message to the IRC server, it was internal only.
*/
func (self *Internal) Special(content string) bool {

	if self.channel == "" {
		// First message from client is either /join or /private,
		// telling us which channel or user this client is talking to

		isPrivate := strings.HasPrefix(content, "/private")
		self.isPrivate = isPrivate

		isJoin := strings.HasPrefix(content, "/join")

		if isJoin || isPrivate {
			parts := strings.Split(content, " ")
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

	if strings.HasPrefix(content, "/notify") {
		self.isNotify = !self.isNotify
		line := Line{
			Command: "NOTICE",
			Content: "Notifications: " + strconv.FormatBool(self.isNotify),
			Channel: "",
			User:    ""}
		jsonData, _ := json.Marshal(line)
		self.netConn.Write(append(jsonData, '\n'))

		// Not an IRC server command, so don't send to IRC server
		return true
	}

	return false
}

func (self *Internal) sendNick() {

	line := Line{
		Command: "NICK",
		Content: self.manager.Nick,
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

	self.manager.fromUser <- Message{self.channel, "/part " + self.channel}
}
