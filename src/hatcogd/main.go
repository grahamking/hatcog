package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
)

const (
	VERSION = "hatcog v0.10 (github.com/grahamking/hatcog)"
)

var (
	host   = flag.String("host", "127.0.0.1", "Internal address to bind")
	port   = flag.String("port", "8790", "Internal port to listen on")
	logdir = flag.String("logdir", "", "Directory for log files")
)

func main() {

	flag.Parse()

	if len(*logdir) != 0 {
		logFilename := *logdir + "/server.log"
		fmt.Println(VERSION, "logging to", logFilename)
		logfile := openLogFile(logFilename)
		log.SetOutput(logfile)
	} else {
		fmt.Println(VERSION, "logging to console")
	}

	log.Println("START")

	server := NewServer(*host, *port)
	defer server.Close()
	go server.Run()

	// Wait for stop signal (Ctrl-C, kill) to exit
	incoming := make(chan os.Signal)
	signal.Notify(incoming, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	for {
		<-incoming
		break
		/*
		   sig := (<-signal.Incoming).(os.UnixSignal)
		   if sig == syscall.SIGINT ||
		       sig == syscall.SIGKILL ||
		       sig == syscall.SIGTERM {
		       break
		   }
		*/
	}
	log.Println("END")
}

// Record panic in log file - run this from defer
func logPanic() {

	if r := recover(); r != nil {
		log.Println("PANIC: ")
		log.Println(r)
		log.Println(string(debug.Stack()))
		panic(r) // Keep right on panicking
	}
}

// Open a file to log to
func openLogFile(logFilename string) *os.File {
	os.Mkdir(*logdir, 0750)

	logFile, err := os.OpenFile(
		logFilename,
		os.O_RDWR|os.O_APPEND|os.O_CREATE,
		0650)
	if err != nil {
		fmt.Println("Error creating log file:", logFilename, err)
		os.Exit(1)
	}
	return logFile
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

// Network string has an optional password. Format is either host:port or host:port:password.
// Split into host:port and password
func splitNetPass(full string) (server, pass string) {

	if strings.Count(full, ":") == 2 {
		parts := strings.Split(full, ":")
		server = strings.Join(parts[0:2], ":")
		pass = parts[2]
	}

	return server, pass
}
