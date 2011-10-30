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
	FULL_NAME          = "Go IRC"
    IRC_NAME_LENGTH    = 15
)

var server = flag.String("server", "127.0.0.1:6667", "IP address or hostname and optional port for IRC server to connect to")
var nick = flag.String("nick", "goirc", "Nick name")
var name = flag.String("name", "Go IRC", "Full name")

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

    // IRC connection to remote server
    var external *Connection
    external = NewConnection(*server, *nick, FULL_NAME, fromServer)
	defer external.Close()

    fmt.Println("Connected to IRC server " + *server)
	go external.Consume()

    // Socket connections from go-join programs
    var internal *Internal
    internal = NewInternal(INTERNAL_PORT, fromUser)
    defer internal.Close()

    fmt.Println("Listening for internal connection on port " + INTERNAL_PORT)
    go internal.Run()

    for isRunning {

        select {
            case serverLine := <-fromServer:
                internal.Write(serverLine.AsJson())

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
        // Input is expected to be '#channel message content ...'
        contentParts := strings.SplitN(content, " ", 2)
        if len(contentParts) != 2 {
            // Invalid message
            return
        }
        channel := contentParts[0]
        content = contentParts[1]
        ircConn.SendMessage(channel, content)
    }

}

// Is 'content' an IRC command?
func isCommand(content string) bool {
	return len(content) > 1 && content[0] == '/'
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

