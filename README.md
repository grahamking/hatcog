goirc is a text-based IRC client which plays well with [tmux](http://www.google.ca/search?q=tmux). I'm using this to learn Go, so you probably shouldn't use it yet. It's still a bit rough. I don't expect it run on anything except Linux.

goirc is made up of two programs: `go-connect`, which connects to your irc server, and `go-join`, which manages input/output for a single channel.

You'll need the [Go language](http://golang.org) installed to be able to use this. Run `gomake` in go-connect/ and in go-join/ to build it.

## Basic usage

In one window start a connection to the server:

    ./go-connect -name="Example Person" -nick="my_nick" -server="irc.freenode.net:6667"

In a different window, join a channel. Note that there's no # in front of the channel name.

    ./go-join test

Finally (and optionally) identify with nickserv by typing this into your go-join window:

    /PRIVMSG NickServ :identify <password>

To join other channels open new windows (panes, in tmux) and join away:

    ./go-join example

To quit a `go-join` just type `/quit`. If you Ctrl-C out of it your terminal will be messy (because we put it into raw mode), so type `stty sane` to reset it.

To actually disconnect from IRC you'll need to Ctrl-C your `go-connect` window - quitting go-join doesn't quit or leave the channel.

## Private messages

goirc's support for private messages is a bit manual at the moment. When you see [PRIVATE] appear in all channels beside a message, you need to get to a new window and run a new go-join:

    ./go-join -private=<nick>

when <nick> is the person who sent you the private message. You now have a special sort-of-channel just to talk to that person.

## Alerts

When someone says your name in a channel or sends you a private message, we try to play a sound and display a notification. You'll need to customise the sound command (in go-connect/main.go). The notification command should work as-is on Ubuntu.

## Supported commands

PRIVMSG, ACTION (/me), QUIT, PART, JOIN, NAMES, NICK

Anything you prefix with / is send direct to the server, so probably many more commands are supported, which don't require special handling.

## Development

A log of raw IRC messages goes to `/tmp/goirc.log`, and a log of go-join's input goes to `/tmp/go-join.log`. For development I use a local install of [ircd-hybrid](https://help.ubuntu.com/community/IrcServer).

Happy Go-IRC-ing!

