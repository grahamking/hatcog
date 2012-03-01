
import os
import sys
import logging
import time
import subprocess
import socket
import random
from Queue import Queue, Empty

from term import Terminal
from remote import Server
from hfilter import translate_in

VERSION    = "hatcog v0.5 (github.com/grahamking/hatcog)"
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
 1. Join channel test: hjoinpy test
 2. Listen for private (/query) message from bob: hjoinpy -private=bob
"""

LOG = logging.getLogger(__name__)

def main(argv=None):
    """Main"""
    if not argv:
        argv = sys.argv

    home = os.path.expanduser('~')
    log_filename = home + LOG_DIR + "clientpy.log"
    print("%s logging to %s" % (VERSION, log_filename))
    logging.basicConfig(filename=log_filename, level=logging.DEBUG)

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

    client = Client(channel, password, DAEMON_ADDR)

    try:
        client.init()
        client.run()
    except:
        LOG.exception("EXCEPTION")
        if client:
            client.stop()

    return 0


class Client(object):
    """Main"""

    def __init__(self, channel, password, daemon_addr):

        self.channel = channel
        self.password = password
        self.daemon_addr = daemon_addr
        self.nick = None

        self.users = UserManager()

        self.from_user = Queue()
        self.terminal = None

        self.from_server = Queue()
        self.server = None

    def init(self):
        """Initialize"""

        self.terminal = Terminal(self.from_user, self.users)
        self.terminal.set_channel(self.channel)

        sock = get_daemon_connection()
        self.server = Server(sock, self.from_server)

        if self.password:
            self.server.write("/pw " + self.password)

        time.sleep(1)
        self.server.write("/join " + self.channel)

    def run(self):
        """Main loop"""

        while 1:
            activity = False

            try:
                msg = self.from_user.get_nowait()
                self.server.write(msg)

                self.terminal.write_msg(self.nick, msg, True)

                activity = True
            except Empty:
                pass

            try:
                msg = self.from_server.get_nowait()
                LOG.debug(msg)
                display = translate_in(msg, self)
                if display:
                    self.terminal.write(display)

                activity = True
            except Empty:
                pass

            if not activity:
                time.sleep(0.1)

    def stop(self):
        """Close remote connection, restore terminal to sanity"""
        if self.terminal:
            self.terminal.stop()
        if self.server:
            self.server.stop()

    #
    # hfilter callbacks
    #

    def on_nick(self, obj, display):
        """A nick change, possibly our own."""
        if not obj['user']:
            self.nick = obj['content']
            self.terminal.set_nick(self.nick)
            return "You are now known as %s" % self.nick

    def on_privmsg(self, obj, display):
        """A message. Format it nicely."""
        username = obj['user']
        self.terminal.write_msg(
                username,
                obj['content'],
                username == self.nick)
        return -1

    def on_join(self, obj, display):
        """User joined channel"""
        self.users.add(obj['user'])
        self.terminal.set_users(self.users.count())

    def on_part(self, obj, display):
        """User left channel"""
        self.users.remove(obj['user'])
        self.terminal.set_users(self.users.count())

    def on_353(self, obj, display):
        """Reply to /names"""
        self.users.add_all(obj['content'])
        self.terminal.set_users(self.users.count())


class UserManager(object):
    """Manages users in an IRC channel"""

    def __init__(self):
        self.users = set()
        self.colors = {}

    def add(self, username):
        """User joined channel"""
        if username.startswith("@") or username.startswith("+"):
            username = username[1:]
        self.users.add(username)

    def remove(self, username):
        """User left channel"""
        self.users.remove(username)

    def add_all(self, usernames):
        """Add a bunch of users"""
        for username in usernames.split(" "):
            self.add(username)

        LOG.debug(self.users)

    def color_for(self, username):
        """The color to display a given user in"""
        if not username in self.colors:
            self.colors[username] = random.choice(range(230))
        return self.colors[username]

    def count(self):
        """Number of users in the channel"""
        return len(self.users)


def get_daemon_connection():
    """Returns a socket connection to the daemon, starting it
    if necessary.
    """

    host, port = DAEMON_ADDR.split(":")

    try:
        sock = socket.create_connection((host, port))
    except:
        sock = start_daemon()

    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)

    return sock


def start_daemon():
    """Start the daemon, and return a connection to it"""
    LOG.debug("Starting the daemon: %s", CMD_START_DAEMON)

    parts = CMD_START_DAEMON.split(" ")
    subprocess.call(parts)

    host, port = DAEMON_ADDR.split(":")

    # Wait for it to be ready
    is_ready = False
    while not is_ready:
        try:
            time.sleep(0.5)
            sock = socket.create_connection((host, port))
            is_ready = True
        except:
            is_ready = False

    return sock


def stop_daemon():
    """Stop the daemon. This will also stop all clients."""
    LOG.debug("Stopping the daemon: %s", CMD_STOP_DAEMON)

    parts = CMD_STOP_DAEMON.split(" ")
    subprocess.call(parts)


if __name__ == '__main__':
    sys.exit(main())
