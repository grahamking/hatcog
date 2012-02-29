
import sys
import logging
from Queue import Queue

import term
from hfilter import translate_in


def main():
    """Main"""

    logging.basicConfig(filename="/tmp/hjoinpy.log", level=logging.DEBUG)
    logging.debug("**** Start")

    from_user = Queue()
    to_user = Queue()
    term.start(from_user, to_user)

    add_msgs(to_user)

    return 0

def add_msgs(to_user):
    """Add IRC messages to queue"""

    max_lines = 10
    current = 0

    for line in open('/home/graham/.hatcog/client_raw.log', 'rt'):
        line = ' '.join(line.split(' ')[3:])
        display = translate_in(line)
        if not display:
            continue

        to_user.put(display)

        current += 1
        if current >= max_lines:
            break


if __name__ == '__main__':
    sys.exit(main())
