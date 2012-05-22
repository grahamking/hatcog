
## Install

Hatcogd is made up of two parts: hatcogd, a server written in Go, and hjoin, a client written in Python. hjoin connects to hatcogd, which connects to the remote IRC server.

1. Copy either `bin/hatcogd-32` (for 32 bit linux) or `bin/hatcogd-64 (for 64 bit linux) onto your path (`/usr/local/bin` is good), and rename it to just `hatcogd`.

    cd /usr/local/bin
    sudo ln -s /home/username/checkout/hatcog/bin/hatcogd-64 hatcogd

 - If on a different system, you'll need to build hatcogd. You'll need [Go](http://golang.org) v1+. Make sure the checkout is on your GOPATH, then type `go build hatcogd`. That will put a `hatcogd` executable in your current directory. Copy or symlink it from `/usr/local/bin`.

1. Symlink hjoin:

     cd /usr/local/bin
     sudo ln -s /home/username/checkout/hatcog/hjoin/hjoin.py hjoin

1. Copy `.hatcogrc` to your home directory. Edit it.

## Run

Run `hjoin <channel>` e.g. hjoin test. There is no hash in front of the channel name.

Log files are in `~/.hatcog/`.

To start a private conversation: `hjoin -private=<nick>`.

The first time (after reboot) you run `hjoin`, it starts the `hatcogd` daemon. When you `/quit` hjoin, the daemon stays running. If you want to kill the daemon, use `hjoin --stop`.

## Details

hatcog is a text-based IRC client which plays well with [tmux](http://www.google.ca/search?q=tmux). The client is in Python / curses, the server in Go.

hatcog is made up of two programs: `hatcogd`, which connects to your irc server, and `hjoin`, which manages input/output for a single channel. `hatcogd` is started for you in the background, so usually you only interact with the curses client, `hjoin`.

The `hjoin` curses interface will display the following information, if available:

 - Top: Server name, time of last ping from server
 - Bottom: Your nick, channel name, channel url, number of active users in channel (spoke in last 10 minutes), number of users in channel.

## Alerts

When someone says your name in a channel or sends you a private message, we try to play a sound and display a notification. You'll need to customise the sound command in .hatcogrc. The notification command should work as-is on Ubuntu.

A common way to display a notification in Ubuntu is using `notify-send`. I prefer to send myself an IM message, that way Pidgin handles displaying it on my desktop and making a sound. It works even if I am using hatcog on a remote machine over ssh. I use `sendxmpp` wrapped in a small bash script.

## Supported commands

See hjoin/hfilter.py for a list. Anything you prefix with / is sent direct to the server. DCC is not supported.

**Standard commands**

 - /me : Display something differently. Try it: `/me eats lunch`.
 - /names : List users in channel.
 - /nick <new_nick> : Change your nickname.
 - /quit : Quit the client. You will part the channel. Server stays running (to stop server `hjoin --stop`).

**Extra commands**

Non-standard IRC commands:

 - /url : Open the most recent url (urls get underlined when displayed) in a browser. Command to open the browser is in .hatcogrc.
 - /notify : Alert me on all messages. Uses the same method of alerting you when someone says your nick, to alert you of every message. Useful for quiet channels, to notice when something happens.
 - /pw : Send your password to identify with NickServ. The client does this for you on startup (password is in .hatcogrc), so you should never need this.

## Development

For development I use a local install of [ircd-hybrid](https://help.ubuntu.com/community/IrcServer), and the 'test' channel on freenode.

Happy IRC-ing!

