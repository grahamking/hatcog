"""Terminal (UI) for hatcog"""

from threading import Thread
import curses
import curses.ascii
from datetime import datetime
import logging
import time

MAX_BUFFER = 100    # Max number of lines to cache for resize / redraw
LOG = logging.getLogger(__name__)


class Terminal(object):
    """Curses user interface"""

    def __init__(self, from_user, user_manager):
        self.from_user = from_user
        self.users = user_manager

        self.stdscr = None
        self.win_header = None
        self.win_output = None
        self.win_input = None
        self.win_status = None

        self.nick = None
        self.cache = {}
        self.lines = []

        self.start()
        self.create_gui()
        self.start_input_thread()

    def start(self):
        """Initialize curses. Copied from curses/wrapper.py"""
        # Initialize curses
        stdscr = curses.initscr()

        # Turn off echoing of keys, and enter cbreak mode,
        # where no buffering is performed on keyboard input
        curses.noecho()
        curses.cbreak()

        # In keypad mode, escape sequences for special keys
        # (like the cursor keys) will be interpreted and
        # a special value like curses.KEY_LEFT will be returned
        stdscr.keypad(1)

        # Start color, too.  Harmless if the terminal doesn't have
        # color; user can test with has_color() later on.  The try/catch
        # works around a minor bit of over-conscientiousness in the curses
        # module -- the error return from C start_color() is ignorable.
        try:
            curses.start_color()
            curses.use_default_colors()
            for i in xrange(1, curses.COLORS):
                curses.init_pair(i, i, -1)
        except:
            LOG.warn("Exception in curses color init")
            pass

        self.stdscr = stdscr

    def stop(self):
        """Stop curses, restore terminal sanity."""
        self.stdscr.keypad(0)
        curses.echo()
        curses.nocbreak()
        curses.endwin()

    def create_gui(self):
        """Create the UI"""
        self.max_height, self.max_width = self.stdscr.getmaxyx()

        self.win_header = self.stdscr.subwin(1, self.max_width, 0, 0)
        self.win_header.bkgdset(" ", curses.A_REVERSE)
        self.win_header.addstr(" " * (self.max_width - 1))
        self.win_header.addstr(0, 0, "+ hatcog +")
        self.win_header.refresh()

        self.win_output = self.stdscr.subwin(
                self.max_height - 3,
                self.max_width,
                1,
                0)
        self.win_output.scrollok(True)
        self.win_output.idlok(True)

        self.win_status = self.stdscr.subwin(
                1,
                self.max_width,
                self.max_height - 2,
                0)
        self.win_status.bkgdset(" ", curses.A_REVERSE)
        self.win_status.addstr(" " * (self.max_width - 1))
        self.win_status.refresh()

        self.stdscr.addstr(self.max_height - 1, 0, "> ")
        self.stdscr.refresh()
        self.win_input = self.stdscr.subwin(
                1,
                self.max_width - 2,
                self.max_height - 1, 2)
        self.win_input.keypad(True)
        # Have to make getch non-blocking, otherwise resize doesn't work right
        self.win_input.nodelay(True)
        self.win_input.refresh()

    def start_input_thread(self):
        """Starts the thread that listen for user input in the textbox"""
        args = (self.win_input, self.from_user, self)
        thread = Thread(name='term', target=input_thread, args=args)
        thread.daemon = True
        thread.start()
        return

    def delete_gui(self):
        """Remove all windows. Called on resize."""

        if self.win_header:
            del self.win_header

        if self.win_output:
            del self.win_output

        if self.win_input:
            del self.win_input

        if self.win_status:
            del self.win_status

        self.stdscr.erase()
        self.stdscr.refresh()

    def write(self, message, refresh=True):
        """Write 'message' to output window"""
        if not message:
            return

        if self.nick in message:            # Highlight nick
            pos = message.find(self.nick)

            before = message[:pos]
            self.win_output.addstr(before)

            self.win_output.addstr(self.nick, curses.A_BOLD)

            after = message[pos + len(self.nick):]
            self.win_output.addstr(after + "\n")

        else:
            self.win_output.addstr(message + "\n")

        if refresh:
            self.lines.append(message)
            # Do "+ 10" here so that we don't slice the buffer on every line
            if len(self.lines) > MAX_BUFFER + 10:
                self.lines = self.lines[len(self.lines) - MAX_BUFFER : ]
            self.win_output.refresh()

    def write_msg(self, username, content, now=None, refresh=True):
        """Write a user message, with fancy formatting"""

        if not now:
            now = datetime.now().strftime("%H:%M")

        self.win_output.addstr(now + " ")

        is_me = username == self.nick
        padded_username = lpad(15, username)
        if is_me:
            self.win_output.addstr(padded_username, curses.A_BOLD)
        else:
            col = self.users.color_for(username)
            self.win_output.addstr(padded_username, curses.color_pair(col))

        self.write(" " + content, refresh=False)

        if refresh:
            self.lines.append((now, username, content))
            self.win_output.refresh()

    def set_nick(self, nick):
        """Set user nick"""
        self.cache['set_nick'] = nick
        self.nick = nick
        self.win_status.addstr(0, 0, nick)
        self.win_status.refresh()

    def set_channel(self, channel):
        """Set current channel"""
        self.cache['set_channel'] = channel
        mid_pos = (self.max_width - (len(channel) + 1)) / 2
        self.win_status.addstr(0, mid_pos, channel, curses.A_BOLD)
        self.win_status.refresh()

    def set_users(self, count):
        """Set number of users"""
        self.cache['set_users'] = count
        msg = "%d users" % count
        right_pos = self.max_width - (len(msg) + 1)
        self.win_status.addstr(0, right_pos, msg)
        self.win_status.refresh()

    def set_host(self, host):
        """Set the host message"""
        self.cache['set_host'] = host
        right_pos = self.max_width - (len(host) + 1)
        self.win_header.addstr(0, right_pos, host)
        self.win_header.refresh()

    def resize(self):
        """Resize the app"""
        self.delete_gui()
        self.create_gui()
        self.start_input_thread()
        self.display_from_cache()

    def display_from_cache(self):
        """GUI was re-drawn, so display all caches data.
        The cache keys are method names, the values
        the arguments to those methods.
        """
        for key, val in self.cache.items():
            getattr(self, key)(val)

        # Now display the main window data
        for line in self.lines:
            if isinstance(line, tuple):
                now, username, content = line
                self.write_msg(username, content, now=now, refresh=False)
            else:
                self.write(line, refresh=False)

        self.win_output.refresh()


def lpad(num, string):
    """Left pad a string"""
    needed = num - len(string)
    return " " * needed + string


#
# Input thread
#

def input_thread(win, from_user, terminal):
    """Listen for user input and write to queue.
    Runs in separate thread.
    """

    while 1:
        try:
            from_user.put( gather_input(win) )
        except ResizeException:
            terminal.resize()
            break


def gather_input(win):
    """Gather input from this window - main input loop"""

    current = []
    pos = 0
    while 1:
        ch = win.getch()
        if ch == -1:
            win.move(0, pos)
            time.sleep(0.1)
            continue

        if ch == curses.KEY_LEFT:
            if pos > 0:
                pos -= 1

        elif ch == curses.KEY_RIGHT:
            if pos < len(current):
                pos += 1

        elif ch == curses.KEY_HOME:
            pos = 0

        elif ch == curses.KEY_END:
            pos = len(current)

        elif ch == curses.KEY_BACKSPACE:
            if pos > 0 and current:
                pos -= 1
                del current[pos]

        elif ch == curses.ascii.NL:     # Enter
            win.erase()
            return ''.join(current)

        elif ch == curses.KEY_RESIZE:
            LOG.debug("window resize")
            # Curses communicates window resize via a fake keypress
            raise ResizeException()

        else:
            # Regular character
            current.insert(pos, chr(ch))
            pos += 1

        redisplay(win, ''.join(current), pos)


def redisplay(win, current, pos):
    win.erase()
    win.addstr(current)
    win.move(0, pos)
    win.refresh()


class ResizeException(Exception):
    """For edit to tell it's thread to quit,
    because we made a new window, so we need a new thread.
    """
    pass

