package main

import (
	"fmt"
	"os"
    "os/signal"
	"strings"
	"bufio"
    "log"
    "syscall"
)

const (
	VERSION    = "hatcog v0.3 (github.com/grahamking/hatcog)"
    DEFAULT_CONFIG = "/.hatcogrc"
    LOG_DIR = "/.hatcog/"
)

var (
    HOME string
    LOG *log.Logger

    fromServer chan *Line
    fromUser chan Message
)


func main() {

    HOME = os.Getenv("HOME")

    logFilename := HOME + LOG_DIR + "main.log"
    fmt.Println(VERSION, "logging to", logFilename)
    openLog(logFilename)

    LOG.Println("START")

    config := loadConfig()
    password := readPassword()

    fromServer = make(chan *Line)
    fromUser = make(chan Message)

    server := NewServer(config, password)
	defer server.Close()
	go server.Run()

    // Wait for stop signal (Ctrl-C, kill) to exit
    for {
        sig := (<-signal.Incoming).(os.UnixSignal)
        if sig == syscall.SIGINT ||
            sig == syscall.SIGKILL ||
            sig == syscall.SIGTERM {
            break
        }
        fmt.Println(sig)
    }
    LOG.Println("END")
}

// Open the main log file
func openLog(logFilename string) {
    os.Mkdir(HOME + LOG_DIR, 0750)

    logFile, err := os.OpenFile(
        logFilename,
        os.O_RDWR|os.O_APPEND|os.O_CREATE,
        0650)
    if err != nil {
        fmt.Println("Error creating log file: "+ logFilename, err)
        os.Exit(1)
    }
    LOG = log.New(logFile, "", log.LstdFlags)
}

// Load / Parse the config file
func loadConfig() Config {

    configFilename := HOME + DEFAULT_CONFIG
    LOG.Println("Reading config file: ", configFilename)

    config, err := Load(configFilename)
    if err != nil {
        fmt.Println("Error parsing config file:", err)
        os.Exit(1)
    }
    return config
}

// Read password from Stdin, if it was piped in. Otherwise return empty string.
func readPassword() string {

    // Stdin is a TTY, not a pipe, so no password
	if isatty(os.Stdin) {
        return ""
    }

    reader := bufio.NewReader(os.Stdin)
    password, _ := reader.ReadString('\n')
    return sane(password)
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
