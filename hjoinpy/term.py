"""Terminal (UI) for hatcog"""

from threading import Thread
import curses
import curses.textpad
from datetime import datetime

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

        self.start()
        self.create_gui()

        args = (self.win_input, self.from_user)
        self.input_thread = Thread(name='term', target=input_thread, args=args)
        self.input_thread.daemon = True
        self.input_thread.start()

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
            for i in xrange(1, curses.COLORS):
                curses.init_pair(i, i, curses.COLOR_BLACK)
        except:
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
        self.stdscr.nodelay(True)

        curses.curs_set(0)

        self.win_header = self.stdscr.subwin(1, self.max_width, 0, 0)
        self.win_header.bkgdset(" ", curses.A_REVERSE)
        self.win_header.addstr(" " * (self.max_width - 2))
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
        self.win_input.refresh()

        # Move cursor to input window
        #curses.setsyx(self.max_height - 1, 2)

    def write(self, message):
        """Write 'message' to output window"""
        if not message:
            return
        self.win_output.addstr(message + "\n")
        self.win_output.refresh()

    def write_msg(self, username, content, is_me):
        """Write a user message, with fancy formatting"""

        now = datetime.now().strftime("%H:%M")
        self.win_output.addstr(now + " ")

        username = lpad(15, username)
        if is_me:
            self.win_output.addstr(username, curses.A_BOLD)
        else:
            col = self.users.color_for(username)
            self.win_output.addstr(username, curses.color_pair(col))

        self.write(" " + content)

    def set_nick(self, nick):
        """Set user nick"""
        self.win_status.addstr(0, 0, nick)
        self.win_status.refresh()

    def set_channel(self, channel):
        """Set current channel"""
        mid_pos = (self.max_width - (len(channel) + 1)) / 2
        self.win_status.addstr(0, mid_pos, channel, curses.A_BOLD)
        self.win_status.refresh()

    def set_users(self, count):
        """Set number of users"""
        msg = "%d users" % count
        right_pos = self.max_width - (len(msg) + 1)
        self.win_status.addstr(0, right_pos, msg)
        self.win_status.refresh()


def input_thread(win, from_user):
    """Listen for user input and write to queue.
    Runs in separate thread.
    """

    textbox = curses.textpad.Textbox(win)
    while True:
        user_input = textbox.edit()
        from_user.put(user_input)
        win.erase()


def lpad(num, string):
    """Left pad a string"""
    needed = num - len(string)
    return " " * needed + string

