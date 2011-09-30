package main

import (
	"flag"
	//"bitbucket.org/binet/go-readline"
)

const (
	INTERNAL_PORT      = "8790"
	FULL_NAME          = "Go IRC"
    IRC_NAME_LENGTH    = 15
)

var server = flag.String("server", "127.0.0.1:6667", "IP address or hostname and optional port for IRC server to connect to")
var nick = flag.String("nick", "goirc", "Nick name")
var channel = flag.String("channel", "#test", "Channel to connect to");


/*
 * main
 */

func main() {

	flag.Parse()
	//flag.Usage()

    var conn *Connection
    conn = NewConnection(*server, *nick, FULL_NAME, *channel)
	defer conn.Close()

    // Internal listener for user input
	go listenInternal(conn)

    // External (IRC server) consume
	conn.Consume()
}

/*
func read() string {
	var ps1 string
	var prompt, line *string

	ps1 = ">> "
	prompt = &ps1

	line = readline.ReadLine(prompt)
	return *line
}
*/
