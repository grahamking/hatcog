package main

import (
	"flag"
	"fmt"
	"os"
	"log"
	"strings"
	"sort"
	"bufio"
)

func init() {
	var logfile *os.File
	logfile, _ = os.Create(RAW_LOG)
	rawLog = log.New(logfile, "", log.LstdFlags)

	infoCmds = sort.StringSlice([]string{"001", "002", "003", "004", "372", "NOTICE"})
	infoCmds.Sort()
}

func main() {

	var password string
	if !isatty(os.Stdin) {
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
