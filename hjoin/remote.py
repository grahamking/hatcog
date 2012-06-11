# coding: utf-8
"""Connection to hatcogd"""

import logging

LOG = logging.getLogger(__name__)


class Server(object):
    """Local proxy for remote hatcogd server"""

    def __init__(self, sock):

        self.conn = sock
        self.data = []

    def write(self, msg):
        """Send a string message to the server"""
        if not msg:
            return
        msg += "\n"
        self.conn.sendall(msg.encode("utf8"))

    def stop(self):
        """Close server connection"""
        self.conn.close()

    def receive_one(self):
        """Listen for data on conn, return it if we have a full line"""

        char = self.conn.recv(1)
        if char == b'\n':
            received = b''.join(self.data)
            self.data = []
            return received.decode("utf8")

        self.data.append(char)
        return None
