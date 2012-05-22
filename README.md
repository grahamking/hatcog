
## Install

Hatcogd is made up of two parts: hatcogd, a server written in Go, and hjoin, a client written in Python. hjoin connects to hatcogd, which connects to the remote IRC server.

1. Build hatcogd. You'll need [Go](http://golang.org) v1+. Make sure the checkout is on your GOPATH, then type `go build hatcogd`. That will put a `hatcogd` executable in your current directory. Copy or symlink it from `usr/local/bin`.

1. Symlink hjoin:

     cd /usr/local/bin
     sudo ln -s /home/username/checkout/hatcog/hjoin/hjoin.py hjoin

1. Copy `.hatcogrc` to your home directory. Edit it.

## Run

Run `hjoin <channel>` e.g. hjoin test

Log files are in `~/.hatcog/`

For a private message: `hjoin -private=<nick>`

## Details

hatcog is a text-based IRC client which plays well with [tmux](http://www.google.ca/search?q=tmux). The client is in Python / curses, the server in Go.

hatcog is made up of two programs: `hatcogd`, which connects to your irc server, and `hjoin`, which manages input/output for a single channel.

## Alerts

When someone says your name in a channel or sends you a private message, we try to play a sound and display a notification. You'll need to customise the sound command in the config file. The notification command should work as-is on Ubuntu.

## Supported commands

See hjoin/hfilter.py for a list. Anything you prefix with / is send direct to the server. DCC is not supported.

## Development

For development I use a local install of [ircd-hybrid](https://help.ubuntu.com/community/IrcServer), and the 'test' channel on freenode.

Happy IRC-ing!

