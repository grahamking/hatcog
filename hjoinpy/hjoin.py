
import sys
import logging
import time
from Queue import Queue, Empty

from term import Terminal
from remote import Server
from hfilter import translate_in


def main():
    """Main"""

    if len(sys.argv) != 2:
        print("Usage: hjoin <channel>")
        return 1

    channel = sys.argv[1]

    logging.basicConfig(filename="/tmp/hjoinpy.log", level=logging.DEBUG)
    logging.debug("**** Start")

    client = Client(channel)
    try:
        client.init()
        client.run()
    except:
        logging.exception("EXCEPTION")
        if client:
            client.stop()

    return 0


class Client(object):
    """Main"""

    def __init__(self, channel):

        self.from_user = Queue()
        self.terminal = None

        self.from_server = Queue()
        self.server = None

        self.channel = channel
        self.nick = None

    def init(self):
        """Initialize"""

        self.terminal = Terminal(self.from_user)
        self.terminal.set_channel(self.channel)

        self.server = Server(self.from_server)
        #self.server.write("/pw " + password)
        time.sleep(1)
        self.server.write("/join #" + self.channel)

    def run(self):
        """Main loop"""

        while 1:
            activity = False

            try:
                msg = self.from_user.get_nowait()
                self.terminal.write(msg)
                self.server.write(msg)

                activity = True
            except Empty:
                pass

            try:
                msg = self.from_server.get_nowait()
                logging.debug(msg)
                self.terminal.write(translate_in(msg, self))

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

    def on_nick(self, obj):
        """A nick change, possibly our own."""
        if not obj['user']:
            self.nick = obj['content']
            self.terminal.set_nick(self.nick)

if __name__ == '__main__':
    sys.exit(main())
