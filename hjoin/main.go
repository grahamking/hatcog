package main

import (
	"fmt"
	"os"
	"log"
	"strings"
	"flag"
    "time"
    "net"
    "exec"
    "bytes"
    "../config"
)

const (
	VERSION    = "hatcog v0.3 (github.com/grahamking/hatcog)"
    DEFAULT_CONFIG = "/.hatcogrc"
    LOG_DIR = "/.hatcog/"

	DAEMON_ADDR        = "127.0.0.1:8790"

    CMD_START_DAEMON = "start-stop-daemon --start --background --exec /usr/local/bin/hatcogd"
    CMD_STOP_DAEMON  = "start-stop-daemon --stop --exec /usr/local/bin/hatcogd"

	ONE_SECOND_NS = 1000 * 1000 * 1000
	RPL_NAMREPLY       = "353"
	RPL_TOPIC          = "332"
	ERR_UNKNOWNCOMMAND = "421"
	CHANNEL_CMDS       = "PRIVMSG, ACTION, PART, JOIN, " + RPL_NAMREPLY

	USAGE         = `
Usage: hjoin [channel|-private=nick]
Note there's no # in front of the channel
Examples:
 1. Join channel test: go-join test
 2. Listen for private (/query) message from bob: go-join -private=bob
`
)

var (
    HOME string
    LOG *log.Logger
)

var userPrivate = flag.String(
	"private",
	"",
	"Listen for private messages from this nick only")

var fromUser = make(chan []byte)
var fromServer = make(chan string)

/*
 * main
 */
func main() {

    HOME = os.Getenv("HOME")

    logFilename := HOME + LOG_DIR + "client.log"
    fmt.Println(VERSION, "logging to", logFilename)
    LOG = openLog(logFilename)

	if len(os.Args) != 2 {
		fmt.Println(USAGE)
		os.Exit(1)
	}

	arg := os.Args[1]
    if arg == "--stop" {
        fmt.Println("Closing all connections")
        stopDaemon()
        return
    }

	var channel string
	if strings.HasPrefix(arg, "-private") {
		flag.Parse()
		channel = *userPrivate
	} else {
		channel = "#" + arg
	}

    conf := loadConfig()
    password := getPassword(conf)

	client := NewClient(channel, password)
	defer func() {
		client.Close()
		fmt.Println("Bye!")
	}()

	client.Run()
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

// Get a connection to the daemon, starting it if needed
func GetDaemonConnection() net.Conn {

    socket, err := net.Dial("tcp", DAEMON_ADDR)
    if err != nil {
        socket = startDaemon()
    }
	socket.SetReadTimeout(ONE_SECOND_NS)
    return socket
}

// Start the daemon, and return a connection to it
func startDaemon() net.Conn {
    LOG.Println("Starting the daemon:", CMD_START_DAEMON)

    parts := strings.Split(CMD_START_DAEMON, " ")
    cmd := exec.Command(parts[0], parts[1:]...)
    cmd.Run()

    // Wait for it to be ready

    var socket net.Conn
    err := os.NewError("Placeholder")

    for err != nil {
        time.Sleep(ONE_SECOND_NS / 2)
        socket, err = net.Dial("tcp", DAEMON_ADDR)
    }
    return socket
}

// Stop the daemon. This will also stop all clients.
func stopDaemon() {
    LOG.Println("Stopping the daemon:", CMD_STOP_DAEMON)

    parts := strings.Split(CMD_STOP_DAEMON, " ")
    cmd := exec.Command(parts[0], parts[1:]...)
    cmd.Run()
}

// Load / Parse the config file
func loadConfig() config.Config {

    configFilename := HOME + DEFAULT_CONFIG
    LOG.Println("Reading config file:", configFilename)

    conf, err := config.Load(configFilename)
    if err != nil {
        fmt.Println("Error parsing config file:", err)
        LOG.Println("Error parsing config file:", err)
        os.Exit(1)
    }
    return conf
}

// Get password from config file
func getPassword(conf config.Config) string {

    password := sane(conf.Get("password"))

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

