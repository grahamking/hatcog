#!/usr/bin/env python3
# coding: utf-8
"""Main part of hatcog IRC client. Run this to start."""

import os
import sys
import logging
import time
import subprocess
import socket
import select
import random

from .term import Terminal
from .remote import Server
from .hfilter import translate_in, is_irc_command
from . import __version__

VERSION = "hatcog v{} (github.com/grahamking/hatcog)".format(__version__)
DEFAULT_CONFIG = "/.hatcogrc"
LOG_DIR = "/.hatcog/"

DAEMON_HOST = "127.0.0.1"
DAEMON_PORT = "8790"

DAEMON_WAIT_SECS = 5

DAEMON = "/usr/local/bin/hatcogd-{arch}"
CMD_START_DAEMON = "start-stop-daemon --start --background --exec {daemon} -- -host={host} -port={port} --logdir {logdir}"
CMD_STOP_DAEMON = "start-stop-daemon --stop --exec {daemon}"

USAGE = """
Usage: hjoin [network.channel|-private=network.nick] [--logger]

Network is one of the keys in the config file: "freenode", "oftc", etc.
There's no # in front of the channel
Examples:
 1. Join channel #test on local dev server:
     hjoin local.test
 2. Join #ubuntu on freenode:
     hjoin freenode.ubuntu
 3. Talk privately (/query) with bob on freenode:
     hjoin -private=freenode.bob

Add "--logger" to act as an IRC logger - gather no input, just print
incoming messages to stdout.

"""

SENSIBLE_AMOUNT = 25

LOG = logging.getLogger(__name__)


def main(argv=None):
    """Main"""
    if not argv:
        argv = sys.argv

    home = os.path.expanduser('~')
    log_filename = home + LOG_DIR + "client.log"
    print("%s logging to %s" % (VERSION, log_filename))
    logging.basicConfig(
            filename=log_filename,
            level=logging.DEBUG,
            format="%(asctime)s %(levelname)s: %(message)s")

    if not len(argv) in (2, 3):
        print(USAGE)
        return 1

    arg = sys.argv[1]
    if arg == "--stop":
        print("Closing all connections")
        stop_daemon()
        return 0

    if "." not in arg:  # Argument must be <network_name>.<channel_name>
        print(USAGE)
        return 1

    network = None
    if arg.startswith("-private="):
        arg = arg[len("-private="):]
        network, channel = arg.split(".")
    else:
        network, channel = arg.split(".")
        channel = "#" + channel

    conf = load_config(os.getenv("HOME"))
    password = get_password(conf)

    host = conf.get("daemon_host", DAEMON_HOST)
    port = conf.get("daemon_port", DAEMON_PORT)
    try:

        client = Client(
                network,
                channel,
                password,
                conf.get("daemon_host", DAEMON_HOST),
                conf.get("daemon_port", DAEMON_PORT),
                conf)

        if len(sys.argv) == 3 and sys.argv[2] == "--logger":
            client = Logger(channel, password, host, port, conf)

        client.init()
        client.run()

    except StopException:
        # Clean exit
        pass
    except:
        LOG.exception("EXCEPTION")
        if client:
            client.stop()
        print("EXCEPTION: See ~{}client.log".format(LOG_DIR))
        show_server_log()

    if client:
        client.stop()

    return 0


class Client(object):
    """Main"""

    def __init__(self,
            network,
            channel,
            password,
            daemon_addr,
            daemon_port,
            conf):

        self.channel = channel
        self.password = password
        self.daemon_addr = daemon_addr
        self.daemon_port = daemon_port

        self.conf = conf
        self.network = network
        try:
            ident_parts = conf[network].split(" ")
        except KeyError:
            print("Network '{}' not found.")
            print("It must be a key in the configuration file.")
            print("Keys found: {}".format(conf.keys()))
            raise StopException()

        self.server_addr = ident_parts[0]
        self.nick = ident_parts[1]
        self.name = " ".join(ident_parts[2:])

        self.users = UserManager()

        self.terminal = None

        self.server = None
        self.is_created = None
        self.is_private = False

        # Support custom /notify command
        self.is_notify = False

    def init(self):
        """Initialize"""

        sock, self.is_created = get_daemon_connection(
                self.daemon_addr, self.daemon_port)
        self.server = Server(sock)

        print("Requesting connection to {}".format(self.server_addr))
        self.server.write("/connect {}".format(self.server_addr))
        time.sleep(2)

        self.start_interface()
        self.start_remote()

    def start_interface(self):
        """Start UI"""

        self.terminal = Terminal(self.users)
        self.terminal.set_channel("{}{}".format(
            self.network, self.channel))

    def start_remote(self):
        """Connect to remote server"""

        if self.is_created:
            # New daemon, introduce ourselves
            self.register()

        self.join()

    def join(self):
        """Join channel or private message"""

        if self.channel.startswith("#"):
            self.server.write("/join " + self.channel)
            time.sleep(1)
        else:
            # Private message (query). /private is not standard.
            self.is_private = True
            self.server.write("/private " + self.channel)

    def register(self):
        """Register ourselves with the server"""

        self.server.write("/nick {}".format(self.nick))
        time.sleep(1)
        self.server.write("/user {nick} 0 * {name}".format(
            nick=self.nick,
            name=self.name))
        time.sleep(1)

        if self.password:
            self.server.write("/pw " + self.password)
            time.sleep(1)

        self.on_nick({"content": self.nick, "user": ""})

    def run(self):
        """Main loop"""

        while 1:

            try:
                ready, _, _ = select.select(
                        [sys.stdin, self.server.conn], [], [])
            except:
                # Window resize signal aborts select
                # Curses makes the resize event a fake keypress, so read it
                self.terminal.receive_one()
                continue

            if self.server.conn in ready:
                sock_data = self.server.receive_one()
                if sock_data:
                    self.act_server(sock_data)

            if sys.stdin in ready:
                term_data = self.terminal.receive_one()
                self.act_user(term_data)

    def act_user(self, msg):
        """Act on user input"""

        if not msg:
            return

        # Local only commands

        if msg == "/quit":
            raise StopException()

        elif msg == "/url":
            url = self.terminal.get_url()
            if url:
                self.terminal.write(url)
                show_url(self.conf, url)
            else:
                self.terminal.write("No url found")
            return

        elif msg == "/notify":
            self.is_notify = not self.is_notify
            self.terminal.write("Notifications: {}".format(self.is_notify))
            return

        # Command massaging

        elif msg.startswith("/msg "):
            # /MSG means a PRIVMSG to a specific user
            msg_cmd, msg_user, msg_content = msg.split(' ', 2)
            msg = "/privmsg {} :{}".format(msg_user, msg_content)

        elif msg == "/names":
            # A blank /names list ALL channels, which is a massive query
            msg = "/names {}".format(self.channel)

        # Send to server

        self.server.write(msg)

        # Output to screen

        if msg.startswith("/me "):
            me_msg = "* " + msg.replace("/me ", self.nick + " ")
            self.terminal.write(me_msg)

        elif not is_irc_command(msg):
            self.terminal.write_msg(self.nick, msg)

    def act_server(self, msg):
        """Act on server data"""

        display = translate_in(msg, self)
        if display:
            self.terminal.write(display)

    def stop(self):
        """Close remote connection, restore terminal to sanity"""
        if self.terminal:
            self.terminal.stop()
        if self.server:
            self.server.stop()

    #
    # hfilter callbacks
    #

    def on_nick(self, obj):
        """A nick change, possibly our own."""

        old_nick = obj['user']
        if old_nick and old_nick not in self.users:
            # User not in our channel
            return -1

        new_nick = obj['content']

        self.users.remove(old_nick)
        self.users.add(new_nick)

        if not old_nick or old_nick == self.nick:
            self.nick = new_nick
            self.terminal.set_nick(self.nick)
            return "You are now known as %s" % self.nick

    def on_privmsg(self, obj):
        """A message. Format it nicely."""
        username = obj['user']
        self.terminal.write_msg(username, obj['content'])

        self.users.mark_active(username)
        self.terminal.set_active_users(self.users.active_count())

        if self.nick in obj['content'] or self.is_notify or self.is_private:
            notify(self.conf, obj)

        return -1

    def on_join(self, obj):
        """User joined channel"""
        self.users.add(obj['user'])
        self.terminal.set_users(self.users.count())
        self.terminal.set_active_users(self.users.active_count())

        # Don't display joins in large channels
        if self.users.count() > SENSIBLE_AMOUNT:
            return -1

    def on_part(self, obj):
        """User left channel"""
        who_left = obj['user']
        if who_left not in self.users:
            # User was not in this channel
            return -1

        self.users.remove(who_left)
        self.terminal.set_users(self.users.count())
        self.terminal.set_active_users(self.users.active_count())

        # Don't display parts in large channels
        if self.users.count() > SENSIBLE_AMOUNT:
            return -1

    def on_quit(self, obj):
        """User quit IRC - treat it the same as leaving the channel"""
        return self.on_part(obj)

    def on_353(self, obj):
        """Reply to /names"""
        self.users.add_all(obj['content'])
        self.terminal.set_users(self.users.count())
        self.terminal.set_active_users(self.users.active_count())

        # Don't display list of users if there's too many
        if self.users.count() > SENSIBLE_AMOUNT:
            return -1

    def on_002(self, obj):
        """Extract host name. The first PING will replace this"""
        host_msg = obj['content'].split(',')[0]
        host_msg = host_msg.replace("Your host is ", "").strip()
        self.terminal.set_host(host_msg)

    def on_328(self, obj):
        """Channel url"""
        url = obj['content']
        msg = "%s (%s)" % (self.channel, url)
        self.terminal.set_channel(msg)
        return -1

    def on_mode(self, obj):
        """Block mode messages with an empty mode"""
        if not obj['content']:
            return -1

    def on_ping(self, obj):
        """Tell the UI we got a server ping"""
        server_name = obj['content']
        self.terminal.set_ping(server_name)
        self.terminal.set_active_users(self.users.active_count())
        return -1

    def on_005(self, obj):
        """Display server settings"""
        settings = ", ".join(obj["args"][1:])
        self.terminal.write("Server settings: {}".format(settings))
        return -1

    def on_451(self, obj):
        """Server demands that we register"""
        self.register()
        return -1

    def on_001(self, obj):
        """RPL_WELCOME, server welcomes us, we can now join a channel."""
        self.join()


class UserManager(object):
    """Manages users in an IRC channel"""

    # How recently does someone have to have spoken to be considered active?
    # Time in seconds (default 10 mins)
    ACTIVE_TIME = 60 * 10

    def __init__(self):
        self.color_choices = [1, 2, 3, 5, 6, 7, 8, 9, 10, 11, 12, 13]

        self.users = set()
        self.colors = {}
        self.last_active = {}

    def __contains__(self, candidate):
        """Support 'in' operate."""
        return candidate in self.users

    def add(self, username):
        """User joined channel"""
        if username.startswith("@") or username.startswith("+"):
            username = username[1:]
        self.users.add(username)

    def remove(self, username):
        """User left channel"""
        try:
            self.users.remove(username)
            del self.last_active[username]
        except KeyError:
            pass

    def add_all(self, usernames):
        """Add a bunch of users"""
        for username in usernames.split(" "):
            self.add(username)

    def color_for(self, username):
        """The color to display a given user in"""
        if not username in self.colors:
            self.colors[username] = random.choice(self.color_choices)
        return self.colors[username]

    def count(self):
        """Number of users in the channel"""
        return len(self.users)

    def first_match(self, nick_part, exclude=None):
        """First nick which starts with 'nick_part'.
        If no match, returns nick_part.
        """
        for nick in self.users:
            exclude_it = exclude and nick in exclude
            if nick.startswith(nick_part) and not exclude_it:
                return nick
        return nick_part

    def mark_active(self, nick):
        """Record activity from given nick"""
        self.last_active[nick] = time.time()

    def active_count(self):
        """Number of active users in the channel"""
        self.purge_last_active()
        return len(self.last_active)

    def purge_last_active(self):
        """Remove inactive users from 'last_active' map"""
        time_ago = time.time() - UserManager.ACTIVE_TIME
        remove = []
        active = self.last_active.copy()
        for nick, last in active.items():
            if last < time_ago:
                remove.append(nick)
        for nick in remove:
            del active[nick]
        self.last_active = active


def load_config(home):
    """Load / Parse the config file"""

    filename = home + DEFAULT_CONFIG
    LOG.debug("Reading config file: %s", filename)

    conf = {}
    try:
        config_file = open(filename)
    except IOError:
        print("Config file not found: {}".format(filename))
        print("Maybe you haven't created it yet.")
        print("Here's one to get you started: " +
              "https://github.com/grahamking/hatcog/blob/master/.hatcogrc")
        sys.exit(1)

    with config_file:

        for line in config_file:
            line = line.strip()
            if not line or line.startswith("#"):
                continue

            key, value = line.split("=")
            conf[key.strip()] = value.strip(" \"'")

    return conf


def get_password(conf):
    """Get password from config file"""

    if not "password" in conf:
        return None

    password = conf["password"].strip()

    if password.startswith("$("):
        pwd_cmd = password[2:len(password) - 1]
        LOG.debug("Running command to get password: %s", pwd_cmd)
        password = subprocess.check_output(pwd_cmd.split(' '))

    return password.decode("utf8")


def show_url(conf, url):
    """Open url in browse"""

    browser = conf["cmd_url"].strip()
    subprocess.Popen(
            [browser, url],
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT)

def notify(conf, obj):
    """Notify user of message, usually using desktop notifications."""

    notifier = conf["cmd_notify"]

    user = obj["user"]
    channel = obj["channel"]

    title = user
    if channel != user:    #Private messages have channel == user
        title += " " + channel

    subprocess.Popen(
            [notifier, title, obj["content"]],
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT)


def get_daemon_connection(host, port):
    """Returns a tuple of a socket connection to the daemon, starting it
    if necessary, and a boolean saying whether we just started the daemon.
    """
    msg = "Connecting to daemon"
    print(msg)
    LOG.info(msg)

    is_created = False

    try:
        sock = socket.create_connection((host.encode("utf8"), int(port)))
    except:
        LOG.exception("Could not connect")
        sock = start_daemon(host, port)
        is_created = True

    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    sock.setblocking(False)     # We're using select, so no need to block

    return (sock, is_created)


def start_daemon(host, port):
    """Start the daemon, and return a connection to it"""

    home = os.path.expanduser('~')
    logdir = home + LOG_DIR

    daemon = DAEMON.format(arch=get_long_size())
    cmd = CMD_START_DAEMON.format(
            daemon=daemon,
            host=host,
            port=port,
            logdir=logdir)
    msg = "Starting daemon: {}".format(cmd)
    print(msg)
    LOG.debug(msg)

    if not os.path.exists(daemon):
        msg = "Daemon not found: {}".format(daemon)
        LOG.error(msg)
        print(msg)
        sys.exit(1)

    parts = cmd.split(" ")
    try:
        out = subprocess.check_output(parts, stderr=subprocess.STDOUT)
        out = out.decode("utf8")
    except subprocess.CalledProcessError:
        msg = "Failed to start daemon"
        print(msg)
        LOG.exception(msg)
        sys.exit(1)

    LOG.debug(out)

    # Wait up to DAEMON_WAIT_SECS for it to be ready
    start_time = time.time()
    is_ready = False
    while not is_ready:
        try:
            time.sleep(0.1)
            sock = socket.create_connection((host, port))
            is_ready = True
        except:
            is_ready = False
            if int(time.time() - start_time) > DAEMON_WAIT_SECS:
                msg = "Failed to start daemon"
                print(msg)
                LOG.error(msg)
                show_server_log()
                sys.exit(1)

    return sock


def show_server_log():
    """Print the last 5 lines of the server log, to help
    users fix server problems.
    """
    slog_filename = os.path.expanduser('~') + LOG_DIR + "server.log"

    # Open server.log as binary because old version of hatcogd logged iso8859
    with open(slog_filename, errors="ignore") as slog:
        last_x = slog.readlines()[-5:]
        print("--- {}".format(slog_filename))
        print("".join(last_x))


def stop_daemon():
    """Stop the daemon. This will also stop all clients."""

    daemon = DAEMON.format(arch=get_long_size())
    cmd = CMD_STOP_DAEMON.format(daemon=daemon)
    LOG.debug("Stopping the daemon: %s", cmd)

    parts = cmd.split(" ")
    out = subprocess.check_output(parts, stderr=subprocess.STDOUT)
    out = out.decode("utf8")
    LOG.debug(out)


def get_long_size():
    """Size of a LONG in bits. Either "32" or "64".
    Used to determine which Go daemon to run.
    """
    try:
        bytestr = subprocess.check_output(["getconf", "LONG_BIT"]).strip()
    except subprocess.CalledProcessError:
        msg = ("Error calling shell command 'getconf LONG_BIT' to determine " +
               "whether system is 32 or 64 bit")
        LOG.exception(msg)
        print(msg)
        sys.exit(1)

    return bytestr.decode("utf8")


class StopException(Exception):
    """Signal that we want to stop the program."""
    pass


class Logger(Client):
    """Just write output to stdout"""

    def start_interface(self):
        self.terminal = None

    def act_server(self, msg):
        """Act on server data"""

        display = translate_in(msg, None, timestamp=True)
        if display:
            print(display)


if __name__ == '__main__':
    sys.exit(main())
