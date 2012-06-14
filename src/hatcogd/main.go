package main

import (
	"exp/terminal"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
    "flag"
)

const (
	VERSION        = "hatcog v0.8 (github.com/grahamking/hatcog)"
	DEFAULT_CONFIG = "/.hatcogrc"
	LOG_DIR        = "/.hatcog/"
)

var (
	HOME string
    host = flag.String("host", "127.0.0.1", "Internal address to bind")
    port = flag.String("port", "8790", "Internal port to listen on")
)

func main() {

    flag.Parse()

	HOME = os.Getenv("HOME")

	if !terminal.IsTerminal(syscall.Stdin) {
		logFilename := HOME + LOG_DIR + "server.log"
		fmt.Println(VERSION, "logging to", logFilename)

		logfile := openLogFile(logFilename)
		log.SetOutput(logfile)

	} else {
		fmt.Println(VERSION, "logging to console")
	}

	log.Println("START")

	conf := loadConfig()
	cmdPrivateChat := conf.Get("cmd_private_chat")

	server := NewServer(*host, *port, cmdPrivateChat)
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

// Open a file to log to
func openLogFile(logFilename string) *os.File {
	os.Mkdir(HOME+LOG_DIR, 0750)

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

// Load / Parse the config file
func loadConfig() Config {

	configFilename := HOME + DEFAULT_CONFIG
	log.Println("Reading config file:", configFilename)

	conf, err := LoadConfig(configFilename)
	if err != nil {
		fmt.Println("Error parsing config file:", err)
		log.Println("Error parsing config file:", err)
		os.Exit(1)
	}
	return conf
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
