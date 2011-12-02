package main

import (
	"fmt"
	"os"
    "os/signal"
	"strings"
    "log"
    "syscall"
    "exec"
    "bytes"
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

    logFilename := HOME + LOG_DIR + "server.log"
    fmt.Println(VERSION, "logging to", logFilename)
    LOG = openLog(logFilename)

    LOG.Println("START")

    config := loadConfig()
    password := getPassword(config)

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
    }
    LOG.Println("END")
}

// Open the main log file
func openLog(logFilename string) *log.Logger {
    os.Mkdir(HOME + LOG_DIR, 0750)

    logFile, err := os.OpenFile(
        logFilename,
        os.O_RDWR|os.O_APPEND|os.O_CREATE,
        0650)
    if err != nil {
        fmt.Println("Error creating log file:", logFilename, err)
        os.Exit(1)
    }
    return log.New(logFile, "", log.LstdFlags)
}

// Load / Parse the config file
func loadConfig() Config {

    configFilename := HOME + DEFAULT_CONFIG
    LOG.Println("Reading config file:", configFilename)

    config, err := Load(configFilename)
    if err != nil {
        fmt.Println("Error parsing config file:", err)
        LOG.Println("Error parsing config file:", err)
        os.Exit(1)
    }
    return config
}

// Get password from config file
func getPassword(config Config) string {

    password := sane(config.Get("password"))

    if strings.HasPrefix(password, "$(") {
        pwdCmd := password[2:len(password)-1]
        LOG.Println("Running command to get password:", pwdCmd)
        password = shell(pwdCmd)
    }
    return password
}

// Run the given command and return it's output
func shell(cmd string) string {

    var stderr, stdout bytes.Buffer

    parts := strings.Split(cmd, " ")
    command := exec.Command(parts[0], parts[1:]...)
    command.Stdout = &stdout
    command.Stderr = &stderr

    err := command.Run()

    if err != nil {
        LOG.Println("Error running command:", err)
        LOG.Println(string(stderr.Bytes()))
        os.Exit(1)
    }

    return string(stdout.Bytes())
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
