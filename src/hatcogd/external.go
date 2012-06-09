package main

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	ONE_SECOND_NS = 1000 * 1000 * 1000 // One second in nanoseconds

	// Standard IRC SSL port
	// http://blog.freenode.net/2011/02/port-6697-irc-via-tlsssl/
	SSL_PORT = "6697"
)

type External struct {
	socket       net.Conn
	isClosing    bool
	fromServer   chan *Line
	rawLog       *log.Logger
	isIdentified bool
}

func NewExternal(server string, fromServer chan *Line) *External {

	logFilename := HOME + LOG_DIR + "server_raw.log"
	rawLog := openLog(logFilename)
	LOG.Println("Logging raw IRC messages to:", logFilename)

	var socket net.Conn
	var err error

	if strings.HasSuffix(server, SSL_PORT) {
		socket, err = tls.Dial("tcp", server, nil)
	} else {
		socket, err = net.Dial("tcp", server)
	}

	if err != nil {
		LOG.Fatal("Error connecting to IRC server:", err)
	}
	time.Sleep(ONE_SECOND_NS)

	conn := External{
		socket:     socket,
		fromServer: fromServer,
		rawLog:     rawLog,
	}

	return &conn
}

// Identify with NickServ. Must of already sent NICK.
func (self *External) Identify(password string) {
	if !self.isIdentified {
		LOG.Println("Identifying with NickServ")
		self.SendMessage("NickServ", "identify "+password)
		self.isIdentified = true
	}
}

// Send a regular (non-system command) IRC message
func (self *External) SendMessage(channel, msg string) {
	fullmsg := "PRIVMSG " + channel + " :" + msg
	self.SendRaw(fullmsg)
}

// Send a /me action message
func (self *External) SendAction(channel, msg string) {
	fullmsg := "PRIVMSG " + channel + " :\u0001ACTION " + msg + "\u0001"
	self.SendRaw(fullmsg)
}

// Send message down socket. Add \n at end first.
func (self *External) SendRaw(msg string) {

	var err error
	msg = msg + "\n"

	self.rawLog.Print(" -->", msg)

	_, err = self.socket.Write([]byte(msg))
	if err != nil {
		LOG.Fatal("Error writing to socket", err)
	}
}

// Process a slash command
func (self *External) doCommand(content string) {

	content = content[1:]
	parts := strings.SplitN(content, " ", 2)
	cmd := parts[0]

	// "msg" is short for "privmsg"
	if cmd == "msg" {
		content = strings.Replace(content, "msg", "privmsg", 1)
	}

	self.SendRaw(content)
}

// Read IRC messages from the connection and act on them
func (self *External) Consume() {

	var contentData []byte
	var content string
	var err error

	bufRead := bufio.NewReader(self.socket)
	for {

		if self.isClosing {
			return
		}

		self.socket.SetReadDeadline(time.Now().Add(ONE_SECOND_NS))
		contentData, err = bufRead.ReadBytes('\n')

		if err != nil {
			netErr, _ := err.(net.Error)

			if netErr.Timeout() == true {
				continue
			} else {
				LOG.Fatal("Consume Error:", err)
			}
		}

		if len(contentData) == 0 {
			continue
		}

		content = toUnicode(contentData)

		self.rawLog.Println(content)

		line, err := ParseLine(content)
		if err == nil {
			self.act(line)
		} else {
			LOG.Println("Invalid line:", content)
		}

	}
}

// Converts an array of bytes to a string
// If the bytes are valid UTF-8, return those (as string),
// otherwise assume we have ISO-8859-1 (latin1, and kinda windows-1252),
// and use the bytes as unicode code points, because ISO-8859-1 is a
// subset of unicode
func toUnicode(data []byte) string {

	var result string

	if utf8.Valid(data) {
		result = string(data)
	} else {
		runes := make([]rune, len(data))
		for index, val := range data {
			runes[index] = rune(val)
		}
		result = string(runes)
	}

	return result
}

// Do something with a line
func (self *External) act(line *Line) {

	if line.Command == "PING" {
		// Reply, and send message on to client
		self.SendRaw("PONG goirc")
	} else if line.Command == "VERSION" {
		versionMsg := "NOTICE " + line.User + " :\u0001VERSION " + VERSION + "\u0001\n"
		self.SendRaw(versionMsg)
	}

	self.fromServer <- line
}

func (self *External) Close() error {
	return self.socket.Close()
}

/* Close connection, return from event loop.
 */
func (self *External) Quit() {
	self.isClosing = true
}
