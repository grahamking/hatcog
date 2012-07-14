# coding: utf-8
"""Terminal (UI) for hatcog"""

import curses
import curses.ascii
from datetime import datetime
import logging
import locale
import tempfile
import subprocess

from .hfilter import is_irc_command

PAGER = ["/usr/bin/less", "+G"]     # Program to display scrollback
MAX_BUFFER = 100    # Max number of lines to cache for resize / redraw
LOG = logging.getLogger(__name__)


class Terminal(object):
    """Curses user interface"""

    def __init__(self, user_manager):
        self.users = user_manager

        self.stdscr = None
        self.win_header = None
        self.win_output = None
        self.win_input = None
        self.win_status = None
        self.term_input = None

        self.nick = None
        self.cache = {}
        self.lines = []

        self.urls = []

        self.user_count = 0
        self.active_user_count = 0

        # File to store what we put on screen, for external scrollback tool
        self.scrollback = tempfile.NamedTemporaryFile()

        self.start()
        self.create_gui()

    def start(self):
        """Initialize curses. Mostly copied from curses/wrapper.py"""

        # This might do some good
        locale.setlocale(locale.LC_ALL, "")

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
            for i in range(1, curses.COLORS):
                curses.init_pair(i, i, -1)
        except:
            LOG.exception("Exception in curses color init")

        self.stdscr = stdscr

    def stop(self):
        """Stop curses, restore terminal sanity."""
        self.stdscr.keypad(0)
        curses.echo()
        curses.nocbreak()
        curses.endwin()

    def get_max_width(self):
        """Width in characters of main screen"""
        _, max_width = self.stdscr.getmaxyx()
        return max_width

    def get_max_height(self):
        """Height in characters of main screen"""
        max_height, _ = self.stdscr.getmaxyx()
        return max_height

    def create_gui(self):
        """Create the UI"""

        max_width = self.get_max_width()
        max_height = self.get_max_height()

        self.win_header = self.stdscr.subwin(1, max_width, 0, 0)
        self.win_header.bkgdset(b" ", curses.A_REVERSE)
        self.win_header.addstr(" " * (max_width - 1))
        if len("hatcog") < max_width:
            self.win_header.addstr(0, 0, "hatcog")
        self.win_header.refresh()

        self.win_output = self.stdscr.subwin(max_height - 3, max_width, 1, 0)
        self.win_output.scrollok(True)
        self.win_output.idlok(True)

        self.win_status = self.stdscr.subwin(1, max_width, max_height - 2, 0)
        self.win_status.bkgdset(b" ", curses.A_REVERSE)
        self.win_status.addstr(" " * (max_width - 1))
        self.win_status.refresh()

        self.stdscr.addstr(max_height - 1, 0, "> ")
        self.stdscr.refresh()
        self.win_input = self.stdscr.subwin(
                1, max_width - 2, max_height - 1, 2)
        self.win_input.keypad(True)
        # Have to make getch non-blocking, otherwise resize doesn't work right
        self.win_input.nodelay(True)
        self.win_input.refresh()

        # Create the input window

        previous_input = ""
        previous_pos = 0
        if self.term_input:
            previous_input = self.term_input.current
            previous_pos = self.term_input.pos

        self.term_input = TermInput(self.win_input, self)
        self.term_input.current = previous_input
        self.term_input.pos = previous_pos

        self.term_input.redisplay()

    def receive_one(self):
        """Main input gather loop"""

        data = None
        try:
            data = self.term_input.gather_one()
        except RebuildException:
            self.rebuild()

        return data

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

        self._write(" ")
        for word in message.split(' '):
            if self.nick and word == self.nick:
                self._write(self.nick, curses.A_BOLD)
            elif word.startswith("http"):
                if not word in self.urls:
                    self.urls.append(word)
                self._write(word, curses.A_UNDERLINE)
            else:
                self._write(word)
            self._write(" ")
        self._write("\n")

        if refresh:
            self.lines.append(message)
            # Do "+ 10" here so that we don't slice the buffer on every line
            if len(self.lines) > MAX_BUFFER + 10:
                self.lines = self.lines[len(self.lines) - MAX_BUFFER:]
            self.win_output.refresh()

            self.term_input.cursor_to_input()

    def write_msg(self, username, content, now=None, refresh=True):
        """Write a user message, with fancy formatting"""

        if not now:
            now = datetime.now().strftime("%H:%M")

        self._write(now + " ")

        is_me = username == self.nick
        padded_username = lpad(15, username)
        if is_me:
            self._write(padded_username, curses.A_BOLD)
        else:
            col = self.users.color_for(username)
            self._write(padded_username, curses.color_pair(col))

        self.write(" " + content, refresh=False)

        if refresh:
            self.lines.append((now, username, content))
            self.win_output.refresh()
            self.term_input.cursor_to_input()

    def _write(self, msg, opt=None):
        """Actually write to output window"""
        if opt:
            self.win_output.addstr(msg, opt)
        else:
            self.win_output.addstr(msg)
        self.scrollback.write(msg.encode("utf8"))

    def set_nick(self, nick):
        """Set user nick"""
        self.cache['set_nick'] = nick

        # Erase previous nick
        if self.nick and len(self.nick) < self.get_max_width():
            self.win_status.addstr(0, 0, " " * len(self.nick))

        # Record and display new nick
        self.nick = nick
        if len(nick) < self.get_max_width():
            self.win_status.addstr(0, 0, nick)
            self.win_status.refresh()

    def set_channel(self, channel):
        """Set current channel"""
        self.cache['set_channel'] = channel
        mid_pos = int((self.get_max_width() - (len(channel) + 1)) / 2)
        if mid_pos > 0:
            self.win_status.addstr(0, mid_pos, channel, curses.A_BOLD)
            self.win_status.refresh()

    def set_users(self, count):
        """Set number of users"""
        self.cache['set_users'] = count
        self.user_count = count - 1         # -1 to not include ourselves
        self._display_user_count()

    def set_active_users(self, active_count):
        """Set number of active users"""
        self.cache['set_active_users'] = active_count
        self.active_user_count = active_count
        self._display_user_count()

    def _display_user_count(self):
        """Display number of users in UI"""
        msg = "{user_count} users ({active_user_count} active)"\
                .format(user_count=self.user_count,
                        active_user_count=self.active_user_count)
        right_pos = self.get_max_width() - (len(msg) + 1)
        if right_pos > 0:   # Skip if window is too narrow
            self.win_status.addstr(0, right_pos, msg)
            self.win_status.refresh()

    def set_host(self, host):
        """Set the host message"""
        if not host or not host.strip():
            return
        self.cache['set_host'] = host
        right_pos = self.get_max_width() - (len(host) + 1)
        if right_pos > 0:
            self.win_header.addstr(0, right_pos, host)
            self.win_header.refresh()

    def set_ping(self, server_name):
        """Received a server ping"""
        now = datetime.now().strftime("%H:%M")
        self.set_host("%s (Last ping %s)" % (server_name, now))

    def rebuild(self):
        """Rebuild the app, usually because it got resized"""
        self.delete_gui()
        self.create_gui()
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
        self.term_input.cursor_to_input()

    def get_url(self):
        """Most recent url"""
        if self.urls:
            return self.urls[-1]
        return None


def lpad(num, string):
    """Left pad a string"""
    needed = num - len(string)
    return " " * needed + string


#
# Input thread
#


class TermInput(object):
    """Gathers input from terminal"""

    KEYS = {
        curses.KEY_LEFT: "key_left",
        curses.KEY_RIGHT: "key_right",
        curses.KEY_HOME: "key_home",
        curses.KEY_END: "key_end",
        curses.KEY_BACKSPACE: "key_backspace",
        curses.ascii.NL: "key_enter",
        curses.KEY_RESIZE: "key_resize",
        curses.KEY_PPAGE: "key_pageup",
        curses.KEY_UP: "key_up",
        curses.KEY_DOWN: "key_down",
        9: "key_tab",  # curses doesn't seem to have a constant
        4: "key_ctrl_d",
    }

    def __init__(self, win, terminal):
        self.win = win
        self.terminal = terminal

        self.current = ""
        self.pending = []   # array of int, partial utf8 characters
        self.pos = 0
        self.previous = ""

        self.cursor_to_input()

    def get_max_len(self):
        """Maximum length of input string (win width - 1)"""
        _, win_input_width = self.terminal.win_input.getmaxyx()
        return win_input_width - 1

    def addstr(self, msg, extra=curses.A_NORMAL):
        """Add string to self.win at current position."""
        try:
            self.win.addstr(msg, extra)
        except curses.error:
            LOG.exception("TermInput: Error adding '%s' with '%s' to display",
                    msg, extra)
            return

    def cursor_to_input(self):
        """Move cursor to input box"""
        move_pos = min(self.pos, self.get_max_len() - 1)
        self.win.move(0, 0)
        self.addstr(self.current[:move_pos])
        self.win.refresh()

    def gather_one(self):
        """Gather single character input from this window"""

        byte = self.win.getch()

        if byte == -1:    # No input
            self.cursor_to_input()

        elif byte in TermInput.KEYS:
            key_func = getattr(self, TermInput.KEYS[byte])
            result = key_func()
            if result:
                return result

        else:
            # Other, either character, or byte of multi-byte char
            self.pending.append(byte)
            try:
                char = bytes(self.pending).decode("utf8")
                self.current = self.current[:self.pos] + char + self.current[self.pos:]
                self.pos += 1
                self.pending = []
            except UnicodeDecodeError:
                # byte is part of multi-byte character
                pass
            except ValueError:
                # byte wasn't valid anything (?)
                self.pending = []

        self.redisplay()

    def redisplay(self):
        """Display current input in input window."""
        self.win.erase()

        msg = self.current

        start = 0
        end = len(msg)
        if len(msg) >= self.get_max_len():
            if self.pos < self.get_max_len():
                end = self.get_max_len() - 1
            else:
                start = min(self.pos, len(msg) - self.get_max_len() + 1)

        extra = curses.A_NORMAL
        if is_irc_command(self.current):
            extra = curses.A_BOLD

        self.addstr(msg[start:end], extra=extra)

        self.win.move(0, 0)
        self.addstr(msg[start:self.pos], extra=extra)
        self.win.refresh()

    def key_left(self):
        """Move one char left"""
        if self.pos > 0:
            self.pos -= 1

    def key_right(self):
        """Move one char right"""
        if self.pos < len(self.current):
            self.pos += 1

    def key_home(self):
        """Move to start of line"""
        self.pos = 0

    def key_end(self):
        """Move to end of line"""
        self.pos = len(self.current)

    def key_backspace(self):
        """Delete char just before cursor"""
        if self.pos > 0 and self.current:
            self.pos -= 1
            self.current = self.current[:self.pos] + self.current[self.pos + 1:]

    def key_enter(self):
        """Send input to queue, clear field"""

        result = self.current
        self.previous = self.current

        self.win.erase()
        self.pos = 0
        self.current = ""

        return result

    def key_resize(self):
        """Curses communicates window resize via a fake keypress"""
        LOG.debug("window resize")
        raise RebuildException()

    def key_tab(self):
        """Auto-complete nick"""

        nick_part = self.word_at_pos(self.current)
        nick = self.terminal.users.first_match(
                nick_part,
                exclude=[self.terminal.nick])

        self.current = self.current.replace(nick_part, nick)
        self.pos += len(nick) - len(nick_part)

    def key_ctrl_d(self):
        return "/quit"

    def word_at_pos(self, current):
        """The word ending at the cursor (pos) in string"""
        if not current:
            return ""

        current = current[:self.pos].strip()
        start = current.rfind(' ')
        if start == -1:
            start = 0
        if start >= self.pos:
            return ""

        return current[start:self.pos].strip()

    def key_pageup(self):
        """PgUp pressed, show scrollback buffer"""
        self.terminal.scrollback.flush()

        cmd = list(PAGER)
        cmd.append(self.terminal.scrollback.name)
        subprocess.call(cmd)

        raise RebuildException()

    def key_up(self):
        """Replace current with previous line"""
        self.current = self.previous
        self.pos = len(self.current)

    def key_down(self):
        """Replace current line with nothing"""
        self.current = ""
        self.pos = 0


class RebuildException(Exception):
    """For edit to tell it's thread to quit,
    because we made a new window, so we need a new thread.
    """
    pass
