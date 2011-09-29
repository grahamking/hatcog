package main

import (
	"strings"
    "os"
    "log"
)

const (
	SYS_COMMANDS = "004 005 254 353 366 376 MODE JOIN PING"
)

type Line struct {
	raw     string
	User    string
	host    string
	Command string
	Args    []string
	Content string
}

var rawLog *log.Logger;
func init() {
    var logfile *os.File;
    logfile, _ = os.Create("/tmp/goirc.log");
    rawLog = log.New(logfile, "", log.LstdFlags);
}

func (self *Line) HasDisplay() bool {
	return *isRaw || !strings.Contains(SYS_COMMANDS, self.Command)
}

func (self *Line) String() string {
	if *isRaw {
		return self.raw
	}

	var output string = ""
	if self.User != "" {
		output += "< " + self.User + "> "
	}
	output += self.Content
	return output
}

func ParseLine(data string) Line {

	var line Line
	var prefix, command, trailing, user, host, raw string
	var args []string = make([]string, 10)
	var parts []string = make([]string, 3)

	data = sane(data)

    rawLog.Println(data);

	raw = data
	if data[0] == ':' { // Do we have a prefix?
		parts = strings.SplitN(data[1:], " ", 2)
		prefix = parts[0]
		data = parts[1]

		if strings.Contains(prefix, "!") {
			parts = strings.Split(prefix, "!")
			user = parts[0]
			host = parts[1]
		} else {
			host = prefix
		}
	}

	if strings.Index(data, " :") != -1 {
		parts = strings.SplitN(data, " :", 2)
		data = parts[0]
		args = strings.Split(data, " ")

		trailing = parts[1]
	} else {
		args = strings.Split(data, " ")
	}
	command = args[0]
	args = args[1:len(args)]

	line = Line{
		raw:     raw,
		User:    user,
		host:    host,
		Command: command,
		Args:    args,
		Content: trailing,
	}

	return line
}
