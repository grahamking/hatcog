package main

import (
	"flag"
    "fmt"
    "os"
    "log"
    "strings"
    "exec"
    "sort"
    "bufio"
)

const (
    VERSION            = "GoIRC v0.2 (github.com/grahamking/goirc)"
    RAW_LOG            = "/tmp/goirc.log"
    NOTIFY_CMD         = "/usr/bin/notify-send"
    SOUND_CMD          = "/usr/bin/aplay -q /home/graham/SpiderOak/xchat_sounds/beep.wav"
)

var serverArg = flag.String("server", "127.0.0.1:6667", "IP address or hostname and optional port for IRC server to connect to")
var nickArg = flag.String("nick", "goirc", "Nick name")
var nameArg = flag.String("name", "Go IRC", "Full name")
var internalPort = flag.String("port", "8790", "Internal port (for go-join)")

var infoCmds sort.StringSlice

var fromServer = make(chan *Line)
var fromUser = make(chan string)

// Logs raw IRC messages
var rawLog *log.Logger;

func init() {
    var logfile *os.File;
    logfile, _ = os.Create(RAW_LOG);
    rawLog = log.New(logfile, "", log.LstdFlags);

    infoCmds = sort.StringSlice([]string{"001", "002", "003", "004", "372", "NOTICE"})
    infoCmds.Sort()
}

/*
 * main
 */
func main() {

    var password string
    if ! isatty(os.Stdin) {
        // Stdin is a pipe, read password
        reader := bufio.NewReader(os.Stdin)
        password, _ = reader.ReadString('\n')
        password = sane(password)
    }

    fmt.Println(VERSION)
    fmt.Println("Logging raw IRC messages to: " + RAW_LOG)

	flag.Parse()

    server := NewServer(*nickArg, *serverArg, *nameArg, *internalPort, password)
    defer server.Close()
    server.Run()

    fmt.Println("Bye")
}

type Server struct {
    nick string
    external *Connection
    internal *Internal
    isRunning bool
}

func NewServer(nick, server, name, internalPort, password string) *Server {

    // IRC connection to remote server
    var external *Connection
    external = NewConnection(server, nick, name, password, fromServer)
    fmt.Println("Connected to IRC server " + server)

    if password != "" {
        fmt.Println("Identifying with NickServ")
    }

    // Socket connections from go-join programs
    var internal *Internal
    internal = NewInternal(internalPort, fromUser, nick)

    fmt.Println("Listening for internal connection on port " + internalPort)

    return &Server{nick, external, internal, false}
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

            case userString := <-fromUser:
                self.onUser(userString)
        }
    }
}

func (self *Server) Close() os.Error {
    self.internal.Close()
    return self.external.Close()
}

// Act on server messages
func (self *Server) onServer(line *Line) {

    if line.Command == "NICK" && line.User == self.nick {
        self.nick = line.Content
        self.internal.Nick = self.nick
        rawLog.Println("Nick change: " + self.nick)
    }

    if contains(infoCmds, line.Command) {
        fmt.Println(line.Content)
    }

    isMsg := (line.Command == "PRIVMSG")
    isPrivateMsg := isMsg && (line.User == line.Channel)

    // Play sound and show notification?
    if (isMsg && strings.Contains(line.Content, self.nick)) || isPrivateMsg {
        self.Notify(line)
    }

    // Send to go-join

    if isPrivateMsg {
        self.internal.WritePrivate(line.AsJson())
        //self.doPrivateMessage(line)
    } else {
        self.internal.Write(line.AsJson())
    }

}

// Act on user input
func (self *Server) onUser(content string) {

    if isCommand(content) {

        /*
        if content == "/quit" {
            isRunning = false
            return
        }
        */

        self.external.doCommand(content)

    } else {
        // Input is expected to be '#channel message_content ...'
        contentParts := strings.SplitN(content, " ", 2)
        if len(contentParts) != 2 {
            // Invalid message
            return
        }
        channel := contentParts[0]
        content = contentParts[1]

        if strings.HasPrefix(content, "/me ") {
            self.external.SendAction(channel, content[4:])
        } else {
            self.external.SendMessage(channel, content)
        }
    }

}

/*
// Open a new window in tmux for the private message.
func (self *Server) doPrivateMessage(line *Line) {
    // TODO: Write this.
    //tmux split-window -v -p 50 'go-join -private=bob'
}
*/

// Alert user that someone is talking to them
func (self *Server) Notify(line *Line) {

    title := line.User
    // Private message have Channel == User so don't repeat it
    if line.Channel != line.User {
        title += " " + line.Channel
    }
    notifyCmd := exec.Command(NOTIFY_CMD, title, line.Content)
    notifyCmd.Run()

    parts := strings.Split(SOUND_CMD, " ")
    soundCmd := exec.Command(parts[0], parts[1:]...)
    soundCmd.Run()
}

// Is 'content' an IRC command?
func isCommand(content string) bool {
	return len(content) > 1 && content[0] == '/'
}

// Does slice 'arr' contain string 'candidate'?
func contains(arr sort.StringSlice, candidate string) bool {
    idx := arr.Search(candidate)
    return idx < len(arr) && arr[idx] == candidate
}

// Is given File a terminal?
func isatty(file *os.File) bool {
    stat, _ := file.Stat()
    return !stat.IsFifo()
}

/* Trims a string to not include junk such as:
- the null bytes after a character return
- \n and \r
- whitespace
- Ascii char \001, which is the extended data delimiter,
  used for example in a /me command before 'ACTION'.
  See http://www.irchelp.org/irchelp/rfc/ctcpspec.html
- Null bytes: \000
*/
func sane(data string) string {
	parts := strings.SplitN(data, "\n", 2)
	return strings.Trim(parts[0], " \n\r\001\000")
}

