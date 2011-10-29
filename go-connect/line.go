package main

import (
    "log"
	"strings"
    "json"
)

const (
	SYS_COMMANDS = "004 005 254 353 366 376 MODE JOIN PING"
)

type Line struct {
	Raw     string
	User    string
	Host    string
	Command string
	Args    []string
	Content string
    IsAction bool
    Channel string
}

func (self *Line) String() string {
    return string(self.AsJson())
}

// Current line as json
func (self *Line) AsJson() []byte {
    jsonData, err := json.Marshal(self)
    if err != nil {
        log.Fatal("Error on json Marshal of " + self.Raw, err)
    }
    // go-join expects lines to have an ending
    jsonData = append(jsonData, '\n')
    return jsonData
}

// Takes a raw string from IRC server and parses it
func ParseLine(data string) *Line {

	var line *Line
	var prefix, command, trailing, user, host, raw string
	var args, parts []string
    var isAction bool

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

		trailing = sane(parts[1])
	} else {
		args = strings.Split(data, " ")
	}
	command = args[0]
	args = args[1:len(args)]

    channel := ""
    for _, arg := range(args) {
        if strings.HasPrefix(arg, "#") {
            channel = arg
            break
        }
    }

    isAction = false
    if strings.HasPrefix(trailing, "ACTION") {
        // Received a /me line
        trailing = strings.SplitN(trailing, " ", 2)[1]
        isAction = true
    }

	line = &Line{
		Raw:     raw,
		User:    user,
		Host:    host,
		Command: command,
		Args:    args,
		Content: trailing,
        IsAction: isAction,
        Channel: channel,
	}

	return line
}

