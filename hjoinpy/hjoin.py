"""Connect to hatcogd (start it if necessary), join an IRC channel,
and provide a user interface to it.
"""

import logging
import sys
import os.path
from client import Client

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

USAGE         = """
Usage: hjoin [channel|-private=nick]
Note there's no # in front of the channel
Examples:
 1. Join channel test: go-join test
 2. Listen for private (/query) message from bob: go-join -private=bob
"""

def main(argv=None):
    if not argv:
        argv = sys.argv

    home = os.path.expanduser('~')
    log_filename = home + LOG_DIR + "clientpy.log"
    print("%s logging to %s" % (VERSION, log_filename))
    logging.basicConfig(filename=log_filename)

    if len(argv) != 2:
        print(USAGE)
        return 1

    arg = sys.argv[1]
    if arg == "--stop":
        print("Closing all connections")
        stop_daemon()
        return 0

    if arg.startswith("-private"):
        print('TODO: private messages')
        #channel = *userPrivate
    else:
		channel = "#" + arg

    #TODO conf = loadConfig()
    #TODO password = getPassword(conf)
    password = ''

    client = Client(channel, password)
    '''
	defer func() {
		client.Close()
		fmt.Println("Bye!")
	}()
    '''

    client.run()


def stop_daemon():
    print("stop_daemon: TODO")
    pass


if __name__ == '__main__':
    sys.exit(main())
