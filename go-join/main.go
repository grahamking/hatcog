package main

import (
	"fmt"
	"os"
	"log"
	"strings"
	"flag"
	"time"
    "exec"
)

const (
	// go-connect and go-join must be on same host for now,
	/// but in future go-connect could be remote
	GO_HOST      = "127.0.0.1:8790"
	RPL_NAMREPLY = "353"
	CHANNEL_CMDS = "PRIVMSG, ACTION, PART, JOIN, " + RPL_NAMREPLY
    //CMD_PRIV_CHAT = "/usr/bin/tmux split-window -v -p 50"
    CMD_PRIV_CHAT = "/usr/bin/gnome-terminal -e"
	USAGE        = `
Usage: go-join [channel|-private=nick]
Note there's no # in front of the channel
Examples:
 1. Join channel test: go-join test
 2. Listen for private (/query) message from bob: go-join -private=bob
`
)

var userPrivate = flag.String(
	"private",
	"",
	"Listen for private messages from this nick only")

var fromUser = make(chan []byte)
var fromServer = make(chan []byte)

// Logs messages from go-connect
var rawLog *log.Logger

func init() {
	var logfile *os.File
	logfile, _ = os.Create("/tmp/go-join.log")
	rawLog = log.New(logfile, "", log.LstdFlags)
}

/*
 * main
 */
func main() {

	if len(os.Args) != 2 {
		fmt.Println(USAGE)
		os.Exit(1)
	}

	var channel string

	arg := os.Args[1]
	if strings.HasPrefix(arg, "-private") {
		flag.Parse()
		channel = *userPrivate
	} else {
		channel = "#" + arg
	}

	client := NewClient(channel)
	defer func() {
		client.Close()
		fmt.Println("Bye!")
	}()

	client.Run()
}

// IRC Client abstraction
type Client struct {
	term      *Terminal
	conn      *InternalConnection
	channel   string
	isRunning bool
	nick      string
	isNames   bool
	users     map[string]bool
}

// Create IRC client. Switch keyboard to raw mode, connect to go-connect socket
func NewClient(channel string) *Client {

	if strings.HasPrefix(channel, "#") {
		fmt.Println("Joining channel " + channel)
	} else {
		fmt.Println("Listening for private messages from " + channel)
	}

	// Set terminal to raw mode, listen for keyboard input
	var term *Terminal = NewTerminal()
	term.Raw()
	term.Channel = channel

	// Connect to go-connect
	var conn *InternalConnection
	conn = NewInternalConnection(GO_HOST, channel)

	return &Client{
		term:    term,
		conn:    conn,
		channel: channel,
		isNames: true, // Accept initial names to fill users map
		users:   make(map[string]bool),
	}
}

/* Main loop
   Listen for keyboard input and socket input and be an IRC client
*/
func (self *Client) Run() {
	var serverData, userInput []byte
	var ok bool

	self.isRunning = true

	go self.conn.Consume()
	go self.term.ListenInternalKeys()

	self.conn.Write([]byte("/names")) // fill users map

	for self.isRunning {

		select {
		case serverData, ok = <-fromServer:
			if ok {
				rawLog.Println(string(serverData))
				self.onServer(serverData)
			} else {
				self.isRunning = false
			}

		case userInput, ok = <-fromUser:
			if ok {
				self.onUser(userInput)
			} else {
				self.isRunning = false
			}
		}
	}

	return
}

// Do something with user input. Usually just send to go-connect
func (self *Client) onUser(content []byte) {

	if sane(string(content)) == "/quit" {
		// Local quit, don't send to server
		// Currently there's no global quit
		self.isRunning = false
		return
	}

	if string(content) == "/names" {
		// Remember we're waiting for name information
		self.isNames = true
	}

	// /me is really a message pretending to be a command,
	isMeCommand := strings.HasPrefix(string(content), "/me")

	if isCommand(content) && !isMeCommand {
		// IRC command
		self.conn.Write(content)

	} else {
		// Send to go-connect
		//self.conn.Write([]byte(self.channel + " "))
		self.conn.Write(content)

		// Display locally

		if isMeCommand {
			content = content[4:]
			self.displayAction(self.nick, string(content))

		} else {
			line := Line{
				Received: time.LocalTime().Format(time.RFC3339),
				User:     self.nick,
				Content:  string(content),
				Channel:  self.channel,
				IsCTCP:   isMeCommand}
			self.term.Write([]byte(line.String(self.nick)))
		}
	}

}

// Do something with Line from go-connect. Usually just display to screen.
func (self *Client) onServer(serverData []byte) {

	line := FromJson(serverData)

	switch line.Command {

	case "PRIVMSG":
        // Is the message a user starting private chat?
        isPrivate := line.Channel == line.User
        if isPrivate && line.Channel != self.channel {
            go self.openPrivate(line.User)
            return
        }

		self.term.Write([]byte(line.String(self.nick)))

	case "ACTION":
		self.displayAction(line.User, line.Content)

	case "JOIN":
		self.display(line.User + " joined the channel")
		self.addUser(line.User)

	case RPL_NAMREPLY:
		if self.isNames {
			self.display("Users currently in " + line.Channel + ": ")
			self.display("  " + line.Content)
			self.isNames = false
			self.updateUsers(strings.Split(line.Content, " "))
		}

	case "NICK":
		if self.nick != "" && !self.users[line.User] {
			return
		}

		if len(line.User) == 0 || line.User == self.nick {
			self.nick = line.Content
			self.display("You are now know as " + self.nick)
		} else {
			self.display(line.User + " is now know as " + line.Content)
		}

		self.removeUser(line.User)
		self.addUser(line.Content)

	case "PART":
		self.display(line.User + " left the channel.")
		self.removeUser(line.User)

	case "QUIT":
		if !self.users[line.User] {
			return
		}
		self.display(line.User + " has quit.")
		self.removeUser(line.User)
	}

}

func (self *Client) addUser(user string) {
	if strings.HasPrefix(user, "@") || strings.HasPrefix(user, "+") {
		user = user[1:]
	}
	self.users[user] = true
}

func (self *Client) removeUser(user string) {
	self.users[user] = false, false
}

func (self *Client) updateUsers(users []string) {
	self.users = make(map[string]bool)
	for _, user := range users {
		self.addUser(user)
	}
}

// Write string to terminal
func (self *Client) display(msg string) {
	self.term.Write([]byte(msg + "\n\r"))
}

// Write an action to the terminal  TODO: This duplicates some of line.String
func (self *Client) displayAction(nick, content string) {
	var formatted string
	if nick == self.nick {
		formatted = Bold(" * " + nick)
	} else {
		formatted = colorfullUser(nick, " * "+nick)
	}

	self.display(formatted + " " + content)
}

// Ask window manager to open a new pane for private messages with given user
func (self *Client) openPrivate(nick string) {
    // TODO: Sanitise nick to prevent command execution
    parts := strings.Split(CMD_PRIV_CHAT, " ")
    parts = append(parts, "go-join -private=" + nick)
    cd := exec.Command(parts[0], parts[1:]...)
    cd.Run()
}

func (self *Client) Close() os.Error {
	self.term.Close()
	return self.conn.Close()
}

// Is 'content' an IRC command?
func isCommand(content []byte) bool {
	return len(content) > 1 && content[0] == '/'
}
