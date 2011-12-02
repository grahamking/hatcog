package main

import (
	"fmt"
	"os"
	"strings"
	"bufio"
)

const (
	VERSION    = "GoIRC v0.3 (github.com/grahamking/goirc)"
    DEFAULT_CONFIG = "/.hatcogrc"
)

var fromServer = make(chan *Line)
var fromUser = make(chan Message)

func main() {

	fmt.Println(VERSION)

    configFilename := os.Getenv("HOME") + DEFAULT_CONFIG
    fmt.Println("Reading config file: " + configFilename)

    config, err := Load(configFilename)
    if err != nil {
        fmt.Println("Error parsing config file:", err)
        return
    }

	var password string
	if !isatty(os.Stdin) {
		// Stdin is a pipe, read password
		reader := bufio.NewReader(os.Stdin)
		password, _ = reader.ReadString('\n')
		password = sane(password)
	}

    server := NewServer(config, password)
	defer server.Close()
	server.Run()

	fmt.Println("Bye")
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
