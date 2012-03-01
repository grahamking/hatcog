"""Connection to hatcogd"""

import socket
from threading import Thread


class Server(object):
    """Local proxy for remote hatcogd server"""

    PORT = 8790

    def __init__(self, from_server):

        self.conn = socket.create_connection(("127.0.0.1", self.PORT))

        args = (self.conn, from_server)
        self.listen_thread = Thread(target=listen_thread, args=args)
        self.listen_thread.daemon = True
        self.listen_thread.start()

    def write(self, msg):
        if not msg:
            return
        self.conn.sendall(msg +"\n")

    def stop(self):
        """Close server connection"""
        self.conn.close()


def listen_thread(conn, queue):
    """Run in separate thread, listen for data on conn, put it on queue"""
    while 1:
        data = []

        char = conn.recv(1)
        while char != "\n":
            data.append(char)
            char = conn.recv(1)

        queue.put(''.join(data))
