/* Switch terminal into raw mode and back. */
package main

import (
    "os"
    "unsafe"
    "syscall"
)

const (
    TTY_FD = 0   // STDIN_FILENO
)

type Terminal struct {
    orig_termios termios
    data []byte
}

func NewTerminal() *Terminal {

    var orig_termios termios
    getTermios(&orig_termios)

    return &Terminal{
        orig_termios: orig_termios,
        data: make([]byte, 1),
    }
}

// Switch to Raw mode (receive each char as pressed)
func (self *Terminal) Raw() os.Error {
    raw := self.orig_termios;

    raw.c_iflag &= ^(BRKINT | ICRNL | INPCK | ISTRIP | IXON);
    raw.c_oflag &= ^(OPOST);
    raw.c_cflag |= (CS8);
    raw.c_lflag &= ^(ECHO | ICANON | IEXTEN | ISIG);

    raw.c_cc[VMIN] = 1;
    raw.c_cc[VTIME] = 0;

    if err := setTermios (&raw); err != nil { return err }

    return nil
}

// Read next single char from Stdin. Blocking.
func (self *Terminal) Read() uint8 {
    bytesRead := 0
    for bytesRead == 0 {
        bytesRead, _ = syscall.Read(TTY_FD, self.data)
    }
    return self.data[0]
}

// Write to Stdout. Implements Writer.
func (self *Terminal) Write(msg []byte) (int, os.Error) {
    bytesWritten, errNo := syscall.Write(TTY_FD, msg)
    err := os.NewError(syscall.Errstr(errNo))
    return bytesWritten, err
}

/*
func (self *Terminal) Writeln(msg string) {
    self.Write(msg)
    self.Write("\n\r")
}
*/

/* Restore terminal settings to what they were at startup.
Implements Closer interface.
*/
func (self *Terminal) Close() os.Error {
    return setTermios(&self.orig_termios)
}

// termios types
type cc_t byte
type speed_t uint
type tcflag_t uint

// termios constants
const (
    BRKINT = tcflag_t (0000002);
    ICRNL = tcflag_t (0000400);
    INPCK = tcflag_t (0000020);
    ISTRIP = tcflag_t (0000040);
    IXON = tcflag_t (0002000);
    OPOST = tcflag_t (0000001);
    CS8 = tcflag_t (0000060);
    ECHO = tcflag_t (0000010);
    ICANON = tcflag_t (0000002);
    IEXTEN = tcflag_t (0100000);
    ISIG = tcflag_t (0000001);
    VTIME = tcflag_t (5);
    VMIN = tcflag_t (6)
)

const NCCS = 32
type termios struct {
    c_iflag, c_oflag, c_cflag, c_lflag tcflag_t;
    c_line cc_t;
    c_cc [NCCS]cc_t;
    c_ispeed, c_ospeed speed_t
}

// ioctl constants
const (
    TCGETS = 0x5401;
    TCSETS = 0x5402
)

func getTermios (dst *termios) os.Error {
    r1, _, errno := syscall.Syscall (syscall.SYS_IOCTL,
                                     uintptr (TTY_FD), uintptr (TCGETS),
                                     uintptr (unsafe.Pointer (dst)));

    if err := os.NewSyscallError ("SYS_IOCTL", int (errno)); err != nil {
        return err
    }

    if r1 != 0 {
        return os.NewError ("Error")
    }

    return nil
}

func setTermios (src *termios) os.Error {
    r1, _, errno := syscall.Syscall (syscall.SYS_IOCTL,
                                     uintptr (TTY_FD), uintptr (TCSETS),
                                     uintptr (unsafe.Pointer (src)));

    if err := os.NewSyscallError ("SYS_IOCTL", int (errno)); err != nil {
        return err
    }

    if r1 != 0 {
        return os.NewError ("Error")
    }

    return nil
}

