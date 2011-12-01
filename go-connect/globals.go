package main

import (
    "flag"
    "sort"
    "log"
)

const (
	VERSION    = "GoIRC v0.2.1 (github.com/grahamking/goirc)"
	RAW_LOG    = "/tmp/goirc.log"
	NOTIFY_CMD = "/usr/bin/notify-send"
	SOUND_CMD  = "/usr/bin/aplay -q /home/graham/SpiderOak/xchat_sounds/beep.wav"
    //PRIV_CHAT_CMD = "/usr/bin/tmux split-window -v -p 50"
    PRIV_CHAT_CMD   = "/usr/bin/gnome-terminal -e"

	ONE_SECOND_NS = 1000 * 1000 * 1000
	RPL_NAMREPLY  = "353"
	PING          = "PING"

    // Standard IRC SSL port
    // http://blog.freenode.net/2011/02/port-6697-irc-via-tlsssl/
    SSL_PORT      = "6697"
)

var serverArg = flag.String("server", "127.0.0.1:6667", "IP address or hostname and optional port for IRC server to connect to")
var nickArg = flag.String("nick", "goirc", "Nick name")
var nameArg = flag.String("name", "Go IRC", "Full name")
var internalPort = flag.String("port", "8790", "Internal port (for go-join)")

var infoCmds sort.StringSlice

var fromServer = make(chan *Line)
var fromUser = make(chan Message)

// Logs raw IRC messages
var rawLog *log.Logger

