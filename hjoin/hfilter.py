# coding: utf-8
"""Map IRC commands to display messages."""

import sys
import json
import logging
from datetime import datetime

PATTERNS_OUT = {

        # Custom command for us to send the password
        '/pw': 'PRIVMSG NickServ :identify %(msg)s',

        # User ACTION
        '/me': 'PRIVMSG %(channel)s :\u0001ACTION %(msg)s\u0001',

        # Any command
        '__default_cmd__': '%(cmd)s %(msg)s',

        # A regular message
        '__default_msg__': 'PRIVMSG %(channel)s :%(msg)s',
}

PATTERNS_IN = {
    'NOTICE': '%(content)s',
    'NICK': '* %(user)s is now known as %(content)s',
    'JOIN': '* %(user)s joined the channel',
    'PART': '* %(user)s left the channel',
    'PRIVMSG': '[%(user)s] \t %(content)s',
    'QUIT': '%(user)s has quit',
    'MODE': 'Mode set to %(content)s',
    'ACTION': '* %(user)s %(content)s',

    # RPL_AWAY
    '301': '%(arg1)s is away: %(content)s',

    # Topic
    '332': 'Topic: %(content)s',

    # NAMES reply
    '353': 'Users in %(channel)s: %(content)s',

    # IRC ops online
    '252': '%(content)s %(arg1)s',

    # Who set the topic
    '333': 'Topic set by %(arg2)s',

    # This channel's URL
    '328': 'Channel url: %(content)s',

    # LINKS reply
    '364': '%(arg1)s: %(content)s',

    # VERSION reply
    '351': 'Server version: %(arg1)s Comment: %(content)s',

    # LIST reply
    '321': '%(arg1)s %(content)s',
    '322': '%(arg1)s %(arg2)s %(content)s',

    '__default__': '%(content)s'
}

IGNORE = [
    '253',  # Num unknown connections
    '254',  # Num channels
    '255',  # Num clients and servers
    '323',  # End of LIST
    '366',  # End of NAMES
    '376',  # End of MOTD
]

DEFAULT = [
    '003',
    '004',
    '265',
    '266',
    '250',
    '251',
    '375'
]

LOG = logging.getLogger(__name__)


def lowercase_keys(dict_obj):
    """Convert map keys to lower case"""
    for key in dict_obj.keys():
        if not key.lower() in dict_obj:
            dict_obj[key.lower()] = dict_obj[key]
            del dict_obj[key]


def add_args(dict_obj):
    """Convert array of args to numbered,
    so that our patterns can access them.
    """

    index = 0
    args = dict_obj['args'] or []
    for item in args:
        dict_obj['arg%d' % index] = item
        index += 1


def translate_in(line, callbacks, timestamp=False):
    """Translate a JSON line from the server
    into display string.

    @param line The line received from server - usually a JSON string.

    @callbacks An object that might have on_<cmd> methods, such as
    on_join, on_privmsg, on_353, etc.
    If that method is found, it is called with the the line as an
    object (the parsed JSON), and the line we would return.
    If that method returns something, we use that instead of the line
    we got from the patterns.
    If that method returns -1, we return nothing. -1 means the callback
    already dealt with things.
    """
    line = line.strip()
    if not line:
        return None

    try:
        obj = json.loads(line)
    except ValueError:
        LOG.exception("JSON parse error on '%s'", line)
        return line

    lowercase_keys(obj)
    cmd = obj['command']
    if cmd in IGNORE:
        return None

    add_args(obj)

    pattern = PATTERNS_IN['__default__']
    if cmd in PATTERNS_IN:
        pattern = PATTERNS_IN[cmd]

    display = pattern % obj

    if timestamp:
        now = datetime.now().isoformat()
        display = now + " " + display

    # Call a method on 'callbacks', for additional processing
    retval = None
    func_name = 'on_' + cmd.lower()
    if callbacks and hasattr(callbacks, func_name):
        retval = getattr(callbacks, func_name)(obj)
        if retval == -1:
            return None

    return retval or display


def translate_out(channel, line):
    """Translate a user input into IRC message."""
    line = line.strip()
    if not line:
        return None

    params = {'channel': channel}

    if is_irc_command(line):
        parts = line[1:].split(' ')
        cmd = parts[0]
        params['cmd'] = cmd

        msg = ' '.join(parts[1:])

        pattern = PATTERNS_OUT['__default_cmd__']
        if cmd in PATTERNS_OUT:
            pattern = PATTERNS_OUT[cmd]

    else:
        # Regular IRC message
        msg = line
        pattern = PATTERNS_OUT['__default_msg__']

    params['msg'] = msg

    output = pattern % params
    return output.encode('utf8')


def is_irc_command(line):
    """Is the given line an IRC command line, as opposed to a message"""
    return line.startswith('/')


# When run as a script, translate everything we see on stdin
if __name__ == '__main__':

    result = None
    if '--in' in sys.argv:
        func_translate = translate_in
    elif '--out' in sys.argv:
        func_translate = lambda x: translate_out(sys.argv[2], x)
    else:
        print('Usage: some_data | hfilter.py [--in|--out] [channel]')
        sys.exit(1)

    for test_line in sys.stdin:

        result = func_translate(test_line)

        if result:
            print(result)

