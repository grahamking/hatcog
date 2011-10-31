package main

import (
	"flag"
    "fmt"
    "os"
    "log"
    "strings"
)

const (
    VERSION            = "0.1"
    RAW_LOG            = "/tmp/goirc.log"
	INTERNAL_PORT      = "8790"
)

var serverArg = flag.String("server", "127.0.0.1:6667", "IP address or hostname and optional port for IRC server to connect to")
var nickArg = flag.String("nick", "goirc", "Nick name")
var nameArg = flag.String("name", "Go IRC", "Full name")

// Logs raw IRC messages
var rawLog *log.Logger;

// Should we keep running?
var isRunning = true

func init() {
    var logfile *os.File;
    logfile, _ = os.Create(RAW_LOG);
    rawLog = log.New(logfile, "", log.LstdFlags);
}

/*
 * main
 */
func main() {

    fromServer := make(chan *Line)
    fromUser := make(chan string)

    fmt.Println("GoIRC v" + VERSION)
    fmt.Println("Logging raw IRC messages to: " + RAW_LOG)

	flag.Parse()

    nick := *nickArg    // go-connect must keep track of our nick

    // IRC connection to remote server
    var external *Connection
    external = NewConnection(*serverArg, nick, *nameArg, fromServer)
	defer external.Close()

    fmt.Println("Connected to IRC server " + *serverArg)
	go external.Consume()

    // Socket connections from go-join programs
    var internal *Internal
    internal = NewInternal(INTERNAL_PORT, fromUser, nick)
    defer internal.Close()

    fmt.Println("Listening for internal connection on port " + INTERNAL_PORT)
    go internal.Run()

    for isRunning {

        select {
            case serverLine := <-fromServer:
                if serverLine.Command == "NICK" && serverLine.User == nick {
                    nick = serverLine.Content
                    internal.Nick = nick
                    rawLog.Println("Nick change: " + nick)
                }
                internal.Write(serverLine.AsJson())

                if serverLine.User == serverLine.Channel {
                    // A private message
                    doPrivateMessage(serverLine)
                }

            case userString := <-fromUser:
                doInput(userString, external)
        }
    }

    fmt.Println("Bye")
}

// Act on user input
func doInput(content string, ircConn *Connection) {

    if isCommand(content) {

        /*
        if content == "/quit" {
            isRunning = false
            return
        }
        */

        ircConn.doCommand(content)

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
            ircConn.SendAction(channel, content[4:])
        } else {
            ircConn.SendMessage(channel, content)
        }
    }

}

// Is 'content' an IRC command?
func isCommand(content string) bool {
	return len(content) > 1 && content[0] == '/'
}

/*
 * Open a new window in tmux for the private message.
 */
func doPrivateMessage(line *Line) {
    // TODO: Write this.
    //tmux split-window -v -p 20 './go-join bob'
}

/* Trims a string to not include junk such as:
- the null bytes after a character return
- \n and \r
- whitespace
- Ascii char \001, which is the extended data delimiter,
  used for example in a /me command before 'ACTION'.
  See http://www.irchelp.org/irchelp/rfc/ctcpspec.html
*/
func sane(data string) string {
	parts := strings.SplitN(data, "\n", 2)
	return strings.Trim(parts[0], " \n\r\001")
}

