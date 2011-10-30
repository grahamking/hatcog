package main

import (
	"strings"
    "json"
    "time"
    "io"
)

const (
	SYS_COMMANDS = "004 005 254 353 366 376 MODE JOIN PING"
)

// Maps nicks to color
var colorMap = make(map[string]string)

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

func (self *Line) HasDisplay() bool {
	return !strings.Contains(SYS_COMMANDS, self.Command)
}

func (self *Line) String() string {

    var now *time.Time
    var output string
    var username string

    now = time.LocalTime()

    // see http://golang.org/src/pkg/time/format.go?s=7285:7328#L17
    output = now.Format("15:04")

	if self.User != "" {

        // TODO: if self.User == conn.nick: username=bold(username)
        username = colorfullUser(self.User)

        if self.IsAction {
            username = Lpad(23, "* " + username)
        } else {
            username = Lpad(23, username)
        }

        output += " " + username + " "
	}

    output += self.Content

    output += "\n\r"
    return output
}

func (self *Line) Display(out io.Writer) {
    out.Write( []uint8(self.String()) )
}

// Take JSON and return a Line
func FromJson(jsonStr []byte) *Line {
    var line *Line = &Line{};
    json.Unmarshal(jsonStr, line)
    return line
}

func colorfullUser(nick string) string {

    if colorMap[nick] == "" {
        nextColorIndex := len(colorMap) % len(UserColors)
        colorMap[nick] = UserColors[nextColorIndex]
    }

    return Color(colorMap[nick], nick)
}
