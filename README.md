
You'd have to be crazy to use this. I'm learning me some Go, so this is just for playing. It does work though.

You'll need go (golang.org) installed to be able to use this. Run 'gomake' to build it.

goirc doesn't do any display management, instead it prints output to stdout, and expects it's input to come from a port. If you use tmux or screen, this can work quite well. Otherwise just open two console windows, one for output and one for input.

In your output window type:
    If you have a local IRC server: ./goirc
    Or to connect to freenode: ./goirc -nick=my_nick -channel=#test -server=irc.freenode.net:6667

In your input window:
    netcat 127.0.0.1 8790

Anything you type into your input window will be sent to the server.

To identify with NickServ, /msg won't work, instead type:
    /PRIVMSG NickServ :identify <password>

Happy Go-IRC-ing!

