package main

import (
	"strings"
    "json"
    "time"
)

const (
	SYS_COMMANDS = "004 005 254 353 366 376 MODE JOIN PING"
)

// Maps nicks to color
var colorMap = make(map[string]string)

type Line struct {
	Raw     string
    Received string
	User    string
	Host    string
	Command string
	Args    []string
	Content string
    IsCTCP  bool
    Channel string
}

func (self *Line) HasDisplay() bool {
	return !strings.Contains(SYS_COMMANDS, self.Command)
}

func (self *Line) String(nick string) string {

    var output, username, content string

    // see http://golang.org/src/pkg/time/format.go?s=7285:7328#L17
    if len(self.Received) != 0 {
        parsedTime, _ := time.Parse(time.RFC3339, self.Received)
        output = parsedTime.Format("15:04") + " "
    }

	if self.User != "" {
        username = self.User

        //if self.Command == "ACTION" {
        //    username = Lpad(15, "* " + username)
        //} else {
        username = Lpad(15, username)
        //}

        if self.User == nick {
            username = Bold(username)
        } else {
            username = colorfullUser(self.User, username)
        }

        output += username + " "
	}

    if self.Channel == self.User {
        output += "[PRIVATE] "
    }

    content = self.Content
    if strings.Contains(content, nick) {
        content = strings.Replace(content, nick, Bold(nick), -1)
    }
    output += content

    output += "\n\r"
    return output
}

// Take JSON and return a Line
func FromJson(jsonStr []byte) *Line {
    var line *Line = &Line{};
    json.Unmarshal(jsonStr, line)
    return line
}

/*
 nick: Nickname of user to look up
 strNick: String to format into color. This usually == nick, but
  can sometimes have a '*' in it for an action.
*/
func colorfullUser(nick string, strNick string) string {

    if colorMap[nick] == "" {
        nextColorIndex := len(colorMap) % len(UserColors)
        colorMap[nick] = UserColors[nextColorIndex]
    }

    return Color(colorMap[nick], strNick)
}
