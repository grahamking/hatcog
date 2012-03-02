#!/usr/bin/env python
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

USAGE         = """
Usage: hjoin [channel|-private=nick]
Note there's no # in front of the channel
Examples:
 1. Join channel test: hjoinpy test
 2. Listen for private (/query) message from bob: hjoinpy -private=bob
"""

SENSIBLE_AMOUNT = 25

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
        channel = arg.split("=")[1]
    else:
		channel = "#" + arg

    conf = load_config(os.getenv("HOME"))
    password = get_password(conf)

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

        if self.channel.startswith("#"):
            self.server.write("/join " + self.channel)
        else:
            # Private message (query). /private is not standard.
            self.server.write("/private " + self.channel)

    def run(self):
        """Main loop"""

        while 1:
            activity = False

            try:
                msg = self.from_user.get_nowait()
                if msg == '/quit':
                    break

                if msg:
                    self.server.write(msg)
                    self.terminal.write_msg(self.nick, msg)
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
        self.terminal.write_msg(username, obj['content'])
        return -1

    def on_join(self, obj, display):
        """User joined channel"""
        self.users.add(obj['user'])
        self.terminal.set_users(self.users.count())

        # Don't display joins in large channels
        if self.users.count() > SENSIBLE_AMOUNT:
            return -1

    def on_part(self, obj, display):
        """User left channel"""
        self.users.remove(obj['user'])
        self.terminal.set_users(self.users.count())

        # Don't display parts in large channels
        if self.users.count() > SENSIBLE_AMOUNT:
            return -1

    def on_quit(self, obj, display):
        """User quit IRC - treat it the same aw leaving the channel"""
        return self.on_part(obj, display)

    def on_353(self, obj, display):
        """Reply to /names"""
        self.users.add_all(obj['content'])
        self.terminal.set_users(self.users.count())

        # Don't display list of users if there's too many
        if self.users.count() > SENSIBLE_AMOUNT:
            return -1

    def on_002(self, obj, display):
        """Extract host name. The first PING will replace this"""
        host_msg = obj['content'].split(',')[0]
        host_msg = host_msg.replace("Your host is ", "").strip()
        self.terminal.set_host(host_msg)

    def on_328(self, obj, display):
        """Channel url"""
        url = obj['content']
        msg = "%s (%s)" % (self.channel, url)
        self.terminal.set_channel(msg)
        return -1

    def on_mode(self, obj, display):
        """Block mode messages with an empty mode"""
        if not obj['content']:
            return -1

    def on_ping(self, obj, display):
        """Tell the UI we got a server ping"""
        server_name = obj['content']
        self.terminal.set_ping(server_name)
        return -1


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

    def color_for(self, username):
        """The color to display a given user in"""
        if not username in self.colors:
            self.colors[username] = random.choice(range(230))
        return self.colors[username]

    def count(self):
        """Number of users in the channel"""
        return len(self.users)


def load_config(home):
    """Load / Parse the config file"""

    filename = home + DEFAULT_CONFIG
    LOG.debug("Reading config file: %s", filename)

    conf = {}
    for line in open(filename):
        line = line.strip()
        if not line or line.startswith("#"):
            continue

        key, value = line.split("=")
        conf[key.strip()] = value.strip(" \"'")

    return conf


def get_password(conf):
    """Get password from config file"""

    password = conf["password"].strip()

    if password.startswith("$("):
        pwd_cmd = password[2:len(password)-1]
        LOG.debug("Running command to get password: %s", pwd_cmd)
        password = subprocess.check_output(pwd_cmd.split(' '))

    return password


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
            time.sleep(0.1)
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
