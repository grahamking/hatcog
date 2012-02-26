
import sys
import json

PATTERNS = {
    'NOTICE': '%(content)s',
    'NICK': '* %(user)s is now known as %(content)s',
    'JOIN': '* %(user)s joined the channel',
    'PRIVMSG': '[%(user)s] \t %(content)s',
    'QUIT': '%(user)s has quit',
    'MODE': 'Mode set to %(content)s',

    # End of MOTD
    '376': '',

    # End of NAMES list
    '366': '',

    # Server info message (?)
    '372': '%(content)s',

    '__default__': '%(content)s'
    }

def lowercase(dict_obj):
    """Convert map keys to lower case"""
    for key in dict_obj.keys():
        if not key.lower() in dict_obj:
            dict_obj[key.lower()] = dict_obj[key]
            del dict_obj[key]

def main():

    missing_cmds = set()

    for line in sys.stdin:
        line = line.strip()
        obj = json.loads(line)
        lowercase(obj)

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
