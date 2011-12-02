package main

import (
	"os"
	"strings"
	"exec"
)

const (
	RPL_NAMREPLY  = "353"
)

var (
    INFO_CMDS = []string{"001", "002", "003", "004", "372", "NOTICE"}
)

type Server struct {
	nick      string
	external  *External
	internal  *InternalManager
	isRunning bool
    cmdNotify string
    cmdBeep string
    cmdPrivateChat string
}

func NewServer(config Config, password string) *Server {

    server := config.Get("server")
    nick := config.Get("nick")
    name := config.Get("name")
    internalPort := config.Get("internal_port")

    cmdNotify := config.Get("cmd_notify")
    cmdBeep := config.Get("cmd_beep")
    cmdPrivateChat := config.Get("cmd_private_chat")

	// IRC connection to remote server
	var external *External
	external = NewExternal(server, nick, name, password, fromServer)
	LOG.Println("Connected to IRC server " + server)

	if password != "" {
		LOG.Println("Identifying with NickServ")
	}

	// Socket connections from go-join programs
	var internal *InternalManager
	internal = NewInternalManager(internalPort, fromUser, nick)

	LOG.Println("Listening for internal connection on port " + internalPort)

	return &Server{nick, external, internal, false, cmdNotify, cmdBeep, cmdPrivateChat}
}

// Main loop
func (self *Server) Run() {
	self.isRunning = true

	go self.external.Consume()
	go self.internal.Run()

	for self.isRunning {

		select {
		case serverLine := <-fromServer:
			self.onServer(serverLine)

		case userMessage := <-fromUser:
			self.onUser(userMessage)
		}
	}
}

func (self *Server) Close() os.Error {
	self.internal.Close()
	return self.external.Close()
}

// Act on server messages
func (self *Server) onServer(line *Line) {

	if isInfoCommand(line.Command) {
		LOG.Println(line.Content)
	}

	if line.Command == "NICK" && line.User == self.nick {
		self.nick = line.Content
		self.internal.Nick = self.nick
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

	if isPrivate || (isMsg && strings.Contains(line.Content, self.nick)) {
		go self.Notify(line)
		go self.Beep()

	} else if isMsg && self.internal.IsNotify(line.Channel) {
		go self.Notify(line)
	}

}

// Act on user input
func (self *Server) onUser(message Message) {

	if isCommand(message.content) {

		self.external.doCommand(message.content)

	} else {

		if strings.HasPrefix(message.content, "/me ") {
			self.external.SendAction(message.channel, message.content[4:])
		} else {
			self.external.SendMessage(message.channel, message.content)
		}
	}

}

// Display this line as an OS notification
func (self *Server) Notify(line *Line) {

	title := line.User
	// Private message have Channel == User so don't repeat it
	if line.Channel != line.User {
		title += " " + line.Channel
	}
	notifyCmd := exec.Command(self.cmdNotify, title, line.Content)
	notifyCmd.Run()
}

// Make a sound to alert user someone is talking to them
func (self *Server) Beep() {
	parts := strings.Split(self.cmdBeep, " ")
	soundCmd := exec.Command(parts[0], parts[1:]...)
	soundCmd.Run()
}

// Ask window manager to open a new pane for private messages with given user
func (self *Server) openPrivate(nick string) {

	// TODO: Sanitise nick to prevent command execution

	parts := strings.Split(self.cmdPrivateChat, " ")
	parts = append(parts, "go-join -private="+nick)

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

// Does slice 'arr' contain string 'candidate'?
/*
func contains(arr sort.StringSlice, candidate string) bool {
	idx := arr.Search(candidate)
	return idx < len(arr) && arr[idx] == candidate
}
*/
