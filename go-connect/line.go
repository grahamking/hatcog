package main

import (
	"os"
	"log"
	"strings"
	"json"
	"time"
)

const (
	SYS_COMMANDS = "004 005 254 353 366 376 MODE JOIN PING"
)

var (
	ELSHORT     = os.NewError("Line too short")
	ELMALFORMED = os.NewError("Malformed line")
)

type Line struct {
	Raw      string
	Received string
	User     string
	Host     string
	Command  string
	Args     []string
	Content  string
	IsCTCP   bool
	Channel  string
}

func (self *Line) String() string {
	return string(self.AsJson())
}

// Current line as json
func (self *Line) AsJson() []byte {
	jsonData, err := json.Marshal(self)
	if err != nil {
		log.Println("Error on json Marshal of "+self.Raw, err)
	}
	// go-join expects lines to have an ending
	jsonData = append(jsonData, '\n')
	return jsonData
}

// Takes a raw string from IRC server and parses it
func ParseLine(data string) (*Line, os.Error) {

	var line *Line
	var prefix, command, trailing, user, host, raw string
	var args, parts []string
	var isCTCP bool

	data = sane(data)
	rawLog.Println(data)

	if len(data) <= 2 {
		return nil, ELSHORT
	}

	raw = data
	if data[0] == ':' { // Do we have a prefix?
		parts = strings.SplitN(data[1:], " ", 2)
		if len(parts) != 2 {
			return nil, ELMALFORMED
		}

		prefix = parts[0]
		data = parts[1]

		if strings.Contains(prefix, "!") {
			parts = strings.Split(prefix, "!")
			if len(parts) != 2 {
				return nil, ELMALFORMED
			}
			user = parts[0]
			host = parts[1]

		} else {
			host = prefix
		}
	}

	if strings.Index(data, " :") != -1 {
		parts = strings.SplitN(data, " :", 2)
		if len(parts) != 2 {
			return nil, ELMALFORMED
		}
		data = parts[0]
		args = strings.Split(data, " ")

		trailing = parts[1]

		// IRC CTCP uses ascii null byte
		if len(trailing) > 0 && trailing[0] == '\001' {
			isCTCP = true
		}
		trailing = sane(trailing)

	} else {
		args = strings.Split(data, " ")
	}

	command = args[0]
	args = args[1:len(args)]

	channel := ""
	for _, arg := range args {
		if strings.HasPrefix(arg, "#") {
			channel = arg
			break
		}
	}

	if len(channel) == 0 {
		if command == "PRIVMSG" {
			// A /query message, fake the user as the channel
			channel = user
		} else if command == "JOIN" {
			// JOIN commands say which channel in content part of msg
			channel = trailing
		}
	}

	if strings.HasPrefix(trailing, "ACTION") {
		// Received a /me line
		parts = strings.SplitN(trailing, " ", 2)
		if len(parts) != 2 {
			return nil, ELMALFORMED
		}
		trailing = parts[1]
		command = "ACTION"
	} else if strings.HasPrefix(trailing, "VERSION") {
		trailing = ""
		command = "VERSION"
	}

	line = &Line{
		Raw:      raw,
		Received: time.LocalTime().Format(time.RFC3339),
		User:     user,
		Host:     host,
		Command:  command,
		Args:     args,
		Content:  trailing,
		IsCTCP:   isCTCP,
		Channel:  channel,
	}

	return line, nil
}
