package main

import (
	"os/exec"
	"strings"
)

const (
	RPL_NAMREPLY = "353"
)

var (
	INFO_CMDS = []string{"001", "002", "003", "004", "372", "NOTICE"}
)

type Server struct {
	nick           string
	external       *External
	internal       *InternalManager
	cmdPrivateChat string
}

func NewServer(conf Config) *Server {

	internalPort := conf.Get("internal_port")

	cmdPrivateChat := conf.Get("cmd_private_chat")

	// Socket connections from client programs
	var internal *InternalManager
	internal = NewInternalManager(internalPort, fromUser)

	LOG.Println("Listening for internal connection on port " + internalPort)

	return &Server{"", nil, internal, cmdPrivateChat}
}

// Main loop
func (self *Server) Run() {

	//go self.external.Consume()
	go self.internal.Run()

	for {

		select {
		case serverLine := <-fromServer:
			self.onServer(serverLine)

		case userMessage := <-fromUser:
			self.onUser(userMessage)
		}
	}
}

func (self *Server) Close() error {
	var err error
	self.internal.Close()
	if self.external != nil {
		err = self.external.Close()
	}
	return err
}

// Act on server messages
func (self *Server) onServer(line *Line) {

	if isInfoCommand(line.Command) {
		LOG.Println(line.Content)
	}

	if len(line.Channel) == 0 && !isChannelRequired(line.Command) {
		self.internal.WriteAll(line.AsJson())
	} else {
		self.internal.WriteChannel(line.Channel, line.AsJson())
	}

	isMsg := (line.Command == "PRIVMSG")
	isPrivate := isMsg && (line.User == line.Channel)

	if isPrivate && !self.internal.HasChannel(line.Channel) {
		self.internal.lastPrivate = []byte(line.AsJson())
		go self.openPrivate(line.User)
	}

}

// Act on user input
func (self *Server) onUser(message Message) {

	var cmd, content string

	if isCommand(message.content) {

		parts := strings.SplitN(message.content[1:], " ", 2)
		cmd = parts[0]
		if len(parts) == 2 {
			content = parts[1]
		}

		if cmd == "pw" {
			self.external.Identify(content)

		} else if cmd == "me" {
			self.external.SendAction(message.channel, content)

		} else if cmd == "nick" {
			newNick := content
			self.nick = newNick
			self.internal.Nick = newNick

			self.external.doCommand(message.content)

		} else if cmd == "connect" {
			// Connect to a remote IRC server

			self.external = NewExternal(content, fromServer)
			LOG.Println("Connected to IRC server", content)
			go self.external.Consume()

		} else {
			self.external.doCommand(message.content)
		}

	} else {
		self.external.SendMessage(message.channel, message.content)
	}

}

// Ask window manager to open a new pane for private messages with given user
func (self *Server) openPrivate(nick string) {

	// TODO: Sanitise nick to prevent command execution

	parts := strings.Split(self.cmdPrivateChat, " ")
	parts = append(parts, "/usr/local/bin/hjoin -private="+nick)

	command := exec.Command(parts[0], parts[1:]...)
	command.Run()
}

// Is 'content' an IRC command?
func isCommand(content string) bool {
	return len(content) > 1 && content[0] == '/'
}

// Is 'command' an IRC information command?
func isInfoCommand(command string) bool {

	for _, cmd := range INFO_CMDS {
		if cmd == command {
			return true
		}
	}
	return false
}

// Does command require a channel
func isChannelRequired(command string) bool {
	return command == RPL_NAMREPLY
}
