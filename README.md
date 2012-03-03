
## Quickstart

Build hatcogd. You'll need [Go](http://golang.org). Type 'gomake' in 'hatcogd'.

Symlink hatcogd and hjoin in /usr/local/bin:
 cd /usr/local/bin
 sudo ln -s /home/username/checkout/hatcog/hjoin/hjoin.py hjoin
 sudo ln -s /home/username/checkout/hatcog/hatcogd/hatcogd hatcogd

Copy .hatcogrc to your home directory. Edit it.

Run 'hjoin.py <channel>' e.g. hjoin test

Log files are in ~/.hatcog/

For a private message: 'hjoin.py -private=<nick>'

## Details

hatcog is a text-based IRC client which plays well with [tmux](http://www.google.ca/search?q=tmux). The client is in Python / curses, the proxy server in Go.

hatcog is made up of two programs: `hatcogd`, which connects to your irc server, and `hjoin`, which manages input/output for a single channel.

You'll need the [Go language](http://golang.org) installed to be able to use this. Run `gomake` in hatcogd/ to build it.

## Alerts

When someone says your name in a channel or sends you a private message, we try to play a sound and display a notification. You'll need to customise the sound command in the config file. The notification command should work as-is on Ubuntu.

## Supported commands

See hjoin/hfilter.py for a list. Anything you prefix with / is send direct to the server. DCC is not supported.

## Development

For development I use a local install of [ircd-hybrid](https://help.ubuntu.com/community/IrcServer), and the 'test' channel on freenode.

Happy IRC-ing!

