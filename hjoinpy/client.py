"""hatcog client - connects to hatcogd
"""

import socket
import multiprocessing
import select


class Client(object):

    def __init__(self, channel, password):

        self.channel = channel
        self.isRunning = False
        self.nick = None

        if channel[0] == "#":
            print("Joining channel %s", channel)
        else:
            print("Listening for private messages from %s", channel)

        #rawLog := openLog(HOME + LOG_DIR + "client_raw.log")

        #self.userManager = UserManager()

        """
        // Set terminal to raw mode, listen for keyboard input
        var term *Terminal = NewTerminal(userManager)
        term.Raw()
        term.Channel = channel
        """

        """ TODO: Send password to hatcogd
        if len(password) != 0:
            socket.Write([]byte("/pw " + password + "\n"))
            time.Sleep(ONE_SECOND_NS)
        """


    def run(self):
        """Main loop"""

        q_sock = multiprocessing.Queue()
        sock_listener = multiprocessing.Process(target=l_socket, args=(q_sock,))
        sock_listener.start()
        print('Started socket listener')

        q_keys = multiprocessing.Queue()
        key_listener = multiprocessing.Process(target=l_keys, args=(q_keys, self.channel))
        key_listener.start()
        print('Started key listener')

        while 1:
            (readable, [], []) = select.select([q_sock._reader, q_keys._reader], [], [])

            for q_ready in readable:
                if q_ready is q_sock._reader:
                    data = q_sock.get()
                elif q_ready is q_keys._reader:
                    data = q_keys.get()

            if data:
                print(data)


        sock_listener.join()
        key_listener.join()


def l_socket(queue):
    """Listen for data from hatcogd socket"""
    sock = socket.create_connection(("127.0.0.1", 8790))
    sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)

    while 1:
        data = sock.recv(4096)
        queue.put(data)

    sock.close()


def l_keys(queue, channel):
    """Listen for keyboard input"""
    while 1:
        data = raw_input("%s " % channel)
        queue.put(data)

