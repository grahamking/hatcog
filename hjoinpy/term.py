"""Terminal (UI) for hatcog"""

from threading import Thread
from Queue import Empty
import curses
import time
import logging


def start(from_user, to_user):

    args = (from_user, to_user)
    thread = Thread(name='term', target=start_thread, args=args)
    thread.start()

def start_thread(from_user, to_user):
    """Runs in a separate thread.
    @param from_user a Queue where we write everything
    the user types.
    @param to_user a Queue where we read and display to the user.
    """
    term = Terminal(from_user, to_user)
    curses.wrapper(term.start)

class Terminal(object):
    """Curses user interface"""

    def __init__(self, from_user, to_user):
        self.from_user = from_user
        self.to_user = to_user

        self.stdscr = None
        self.win_status = None
        self.win_input = None
        self.win_output = None

        self.input_pos = 1
        self.input_str = ""

    def start(self, stdscr):
        """Create the UI"""
        self.stdscr = stdscr
        max_height, max_width = self.stdscr.getmaxyx()
        self.stdscr.nodelay(True)

        self.win_status = stdscr.subwin(1, max_width, 0, 0)
        self.win_status.bkgdset(" ", curses.A_REVERSE)
        self.win_status.addstr(" " * (max_width - 1))
        self.win_status.addstr(0, 0, "Status bar goes here")
        self.win_status.refresh()

        self.win_input = stdscr.subwin(3, max_width - 1, 1, 0)
        self.win_input.border()
        self.win_input.refresh()

        self.win_output = stdscr.subwin(max_height - 4, max_width, 4, 0)
        self.win_output.scrollok(True)
        self.win_output.idlok(True)

        #textbox.edit()

        self.main_loop()

    def main_loop(self):
        """Listen for user input or to_user queue input"""
        while True:
            activity = False

            try:
                display = self.to_user.get_nowait()
                self.win_output.addstr(display + "\n")
                self.win_output.refresh()
                activity = True
            except Empty:
                # Nothing to display yet
                pass

            char = self.stdscr.getch()
            if char != -1:
                activity = True

                if char == 10:
                    logging.debug('ENTER')
                    self.from_user.put(self.input_str)
                    self.input_pos = 1
                    self.input_str = ""
                    self.win_input.erase()
                else:
                    logging.debug(char)
                    self.win_input.addch(1, self.input_pos, char)
                    self.input_pos += 1
                    self.input_str += str(char)

                self.win_input.refresh()

            if not activity:
                time.sleep(0.1)

