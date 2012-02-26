
import sys
import json
from datetime import datetime

PATTERNS = {
    'NOTICE': u'%(content)s',
    'NICK': u'* %(user)s is now known as %(content)s',
    'JOIN': u'* %(user)s joined the channel',
    'PART': u'* %(user)s left the channel',
    'PRIVMSG': u'[%(user)s] \t %(content)s',
    'QUIT': u'%(user)s has quit',
    'MODE': u'Mode set to %(content)s',

    # Message of the day
    '372': u'%(content)s',

    # Topic
    '332': u'Topic: %(content)s',

    # NAMES reply
    '353': u'Users currently in %(channel)s: %(content)s',

    # IRC ops online
    '252': u'%(content)s %(arg1)s',

    # Who set the topic
    '333': u'Topic set by %(arg2)s',

    # This channel's URL
    '328': u'Channel url: %(content)s',

    '__default__': u'%(content)s'
}

IGNORE = [
    '005',  # Extensions server supports
    '253',  # Num unknown connections
    '254',  # Num channels
    '255',  # Num clients and servers
    '366',  # End of NAMES
    '376',  # End of MOTD
]

DEFAULT = [
    '001',
    '002',
    '003',
    '004',
    '265',
    '266',
    '250',
    '251',
    '375'
]

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


def main():

    missing_cmds = set()

    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue

        obj = json.loads(line)
        lowercase_keys(obj)
        cmd = obj['command']
        if cmd in IGNORE:
            continue

        add_args(obj)

        if cmd in PATTERNS:
            pattern = PATTERNS[cmd]
        elif cmd in DEFAULT:
            pattern = PATTERNS['__default__']
        else:
            missing_cmds.add(cmd)
            # Use default setup because we didn't account for this message
            pattern = PATTERNS['__default__']

        # Timestamp everything
        if '--timestamp' in sys.argv:
            pattern = '%(received)s ' + pattern

        output = pattern % obj
        print(output.encode('utf8'))

    if missing_cmds:
        print(missing_cmds)


if __name__ == '__main__':
    sys.exit(main())
