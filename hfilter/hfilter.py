
import sys
import json
from datetime import datetime

PATTERNS = {
    'NOTICE': '%(content)s',
    'NICK': '* %(user)s is now known as %(content)s',
    'JOIN': '* %(user)s joined the channel',
    'PRIVMSG': '%(time)s [%(user)s] \t %(content)s',
    'QUIT': '%(user)s has quit',
    'MODE': 'Mode set to %(content)s',

    # Message of the day
    '372': '%(content)s',

    # Topic
    '332': 'Topic: %(content)s',

    # NAMES reply
    '353': 'Users currently in %(channel)s: %(content)s',

    # IRC ops online
    '252': '%(arg1)s %(content)s',

    # Who set the topic
    '333': 'Topic set by %(arg2)s',

    # This channel's URL
    '328': 'Channel url: %(content)s',

    # Ignores
    '005': '',  # Extensions server supports
    '253': '',  # Num unknown connections
    '254': '',  # Num channels
    '255': '',  # Num clients and servers
    '366': '',  # End of NAMES
    '376': '',  # End of MOTD

    # 001, 002, 003,
    # 004, 265, 266,
    # 250, 251, 375
    '__default__': '%(content)s'
}


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
    for item in dict_obj['args']:
        dict_obj['arg%d' % index] = item
        index += 1


def add_time(dict_obj):
    """Add HH:mm under key 'time' to dict"""
    #received = datetime.strpdict_obj['received']
    dict_obj['time'] = datetime.now().strftime('%H:%M')


def main():

    missing_cmds = set()

    for line in sys.stdin:
        line = line.strip()
        obj = json.loads(line)
        lowercase_keys(obj)
        add_args(obj)
        add_time(obj)

        cmd = obj['command']

        if cmd in PATTERNS:
            pattern = PATTERNS[cmd]
        else:
            pattern = PATTERNS['__default__']
            missing_cmds.add(cmd)

        print(pattern % obj)

    print(missing_cmds)


if __name__ == '__main__':
    sys.exit(main())
