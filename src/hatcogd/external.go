package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"net"
	"time"
	"unicode/utf8"
)

const (
	ONE_SECOND_NS = 1000 * 1000 * 1000 // One second in nanoseconds
)

/*******************
 * ExternalManager *
 *******************/

type ExternalManager struct {
	connections map[string]*External
	fromServer  chan *Line
}

func NewExternalManager(fromServer chan *Line) *ExternalManager {
	return &ExternalManager{make(map[string]*External), fromServer}
}

func (self *ExternalManager) Connect(addr string) {

	server, pass := splitNetPass(addr)

	if self.connections[server] == nil {
		self.connections[server] = NewExternal(server, pass, self.fromServer)
		go self.connections[server].Consume()
	}
}

func (self *ExternalManager) Identify(network, password string) {
	ext := self.connections[network]
	if ext == nil {
		log.Println("Error: no network for ", network)
		return
	}
	ext.Identify(password)
}

func (self *ExternalManager) SendMessage(network, channel, msg string) {
	ext := self.connections[network]
	if ext == nil {
		log.Println("Error: no network for ", network)
		return
	}
	ext.SendMessage(channel, msg)
}

func (self *ExternalManager) SendAction(network, channel, msg string) {
	ext := self.connections[network]
	if ext == nil {
		log.Println("Error: no network for ", network)
		return
	}
	ext.SendAction(channel, msg)
}

func (self *ExternalManager) doCommand(network, content string) {
	ext := self.connections[network]
	if ext == nil {
		log.Println("Error: no network for ", network)
		return
	}
	ext.doCommand(content)
}

func (self *ExternalManager) Close() error {
	for _, conn := range self.connections {
		conn.Close()
	}
	self.connections = nil
	return nil
}

/************
 * External *
 ************/

type External struct {
	network      string
	pass         string
	socket       net.Conn
	fromServer   chan *Line
	rawLog       *log.Logger
	isIdentified bool
}

func NewExternal(server string, pass string, fromServer chan *Line) *External {

	logFilename := *logdir + "/server_raw.log"
	logFile := openLogFile(logFilename)
	rawLog := log.New(logFile, "", log.LstdFlags)
	log.Println("Logging raw IRC messages to:", logFilename)

	conn := &External{
		network:    server,
		pass:       pass,
		fromServer: fromServer,
		rawLog:     rawLog,
	}
	conn.connect()

	return conn
}

func (self *External) connect() {

	var err error

	self.socket, err = sock(self.network, 5)
	if err != nil {
		log.Fatal("Error connecting to IRC server: ", err)
	}

	time.Sleep(ONE_SECOND_NS)

	if self.pass != "" {
		self.SendRaw("PASS " + self.pass)
	}
}

/* A socket connection to give network (ip:port). */
func sock(network string, tries int) (net.Conn, error) {

	var socket net.Conn
	var err error

	for tries > 0 {

		socket, err = tls.Dial("tcp", network, nil) // Always try TLS first
		if err == nil {
			log.Println("Secure TLS connection to", network)
			return socket, nil
		}

		socket, err = net.Dial("tcp", network)
		if err == nil {
			log.Println("Insecure connection to", network)
			return socket, nil
		}

		log.Println("Connection attempt failed:", err)
		time.Sleep(ONE_SECOND_NS)

		tries--
	}

	return socket, err
}

// Identify with NickServ. Must of already sent NICK.
func (self *External) Identify(password string) {
	if !self.isIdentified {
		log.Println("Identifying with NickServ")
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
	if err == io.EOF {
		log.Println("SendRaw: IRC server closed connection.")
		self.Close()
	} else if err != nil {
		log.Fatal("Error writing to socket: ", err)
	}
}

// Process a slash command
func (self *External) doCommand(content string) {

	content = content[1:]
	self.SendRaw(content)
}

// Read IRC messages from the connection and act on them
func (self *External) Consume() {
	defer logPanic()

	var contentData []byte
	var content string
	var err error

	bufRead := bufio.NewReader(self.socket)
	for {

		self.socket.SetReadDeadline(time.Now().Add(ONE_SECOND_NS))
		contentData, err = bufRead.ReadBytes('\n')

		if err != nil {
			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() == true {
				continue
			} else if err == io.EOF {
				log.Println("Consume: IRC server closed connection.")
				self.Close()

				// Reconnect
				log.Println("Attempting to reconnect")
				self.connect()
				bufRead = bufio.NewReader(self.socket)
				continue

			} else {
				log.Fatal("Consume Error:", err)
			}
		}

		if len(contentData) == 0 {
			continue
		}

		content = toUnicode(contentData)

		self.rawLog.Println(content)

		line, err := ParseLine(content)
		if err == nil {
			line.Network = self.network
			self.act(line)
		} else {
			log.Println("Invalid line:", content)
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
		self.SendRaw("PONG " + line.Content)
	} else if line.Command == "VERSION" {
		versionMsg := "NOTICE " + line.User + " :\u0001VERSION " + VERSION + "\u0001\n"
		self.SendRaw(versionMsg)
	}

	self.fromServer <- line
}

func (self *External) Close() error {
	return self.socket.Close()
}
