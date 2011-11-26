/* Switch terminal into raw mode and back. */
package main

import (
	"os"
	"unsafe"
	"syscall"
	"strconv"
)

const (
	TTY_FD      = 0 // STDIN_FILENO
	_TIOCGWINSZ = 0x5413
	_TIOCSWINSZ = 0x5414
)

// Used by GetWinsize
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

type Terminal struct {
	orig_termios termios
	data         []byte
	Channel      string
	input        []byte
	cursorPos    int // Cursor position
}

func NewTerminal() *Terminal {

	var orig_termios termios
	getTermios(&orig_termios)

	return &Terminal{
		orig_termios: orig_termios,
		data:         make([]byte, 1),
		input:        make([]byte, 0),
	}
}

// Switch to Raw mode (receive each char as pressed)
func (self *Terminal) Raw() os.Error {
	raw := self.orig_termios

	raw.c_iflag &= ^(BRKINT | ICRNL | INPCK | ISTRIP | IXON)
	raw.c_oflag &= ^(OPOST)
	raw.c_cflag |= (CS8)
	raw.c_lflag &= ^(ECHO | ICANON | IEXTEN | ISIG)

	raw.c_cc[VMIN] = 1
	raw.c_cc[VTIME] = 0

	if err := setTermios(&raw); err != nil {
		return err
	}

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
	self.ClearLine()
	bytesWritten, err := self.rawWrite(msg)
	self.displayInput()
	return bytesWritten, err
}

// Output bytes to Stdout
func (self *Terminal) rawWrite(msg []byte) (int, os.Error) {
	bytesWritten, errNo := syscall.Write(TTY_FD, msg)
	err := os.NewError(syscall.Errstr(errNo))
	return bytesWritten, err
}

// Clear current line by writing blanks and a \r
func (self *Terminal) ClearLine() {
	self.rawWrite([]byte("\r"))
	for i := 0; i < TerminalWidth(); i++ {
		self.rawWrite([]byte(" "))
	}
	self.rawWrite([]byte("\r"))
}

/* Restore terminal settings to what they were at startup.
   Implements Closer interface.
*/
func (self *Terminal) Close() os.Error {
	return setTermios(&self.orig_termios)
}

// Listen for keypresses, display them
func (self *Terminal) ListenInternalKeys() {

	var char byte

	for {
		char = self.Read()

		if char == 0x09 {
			// Find previous space
			// Match prefix against nick list
			// Replace from previous space with match
		}

		if char == 0x7f && len(self.input) > 0 && self.cursorPos > 0 {
			// 0x7f = Unicode Backspace
			if self.cursorPos == len(self.input) {
				self.input = self.input[:len(self.input)-1]
			} else {
				// Delete
				self.input = append(self.input[:self.cursorPos-1], self.input[self.cursorPos:]...)
			}
			self.cursorPos -= 1
		}

		if char == 0x1B {
			// ESC code - starts escape sequence
			char = self.Read()

            // '[' is ANSI escape, 0x4f comes before Home and End
			if ! (char == '[' || char == 0x4F) {
                rawLog.Println("Unexpected char after ESC", char)
                continue
            }

            char = self.Read()
            switch char {

            case 'D':   // Arrow left
                self.cursorPos -= 1
                if self.cursorPos < 0 {
                    self.cursorPos = 0
                }

            case 'C':   // Arrow right
                self.cursorPos += 1
                if self.cursorPos > len(self.input) {
                    self.cursorPos = len(self.input)
                }

            case 'H':   // Home
                self.cursorPos = 0

            case 'F':   // End
                self.cursorPos = len(self.input)

            default:
                rawLog.Println("Unknown escape sequence:", char)
            }

		} else if char >= 0x20 && char < 0x7f {
			// Only use printable characters. See 'man ascii'
			if self.cursorPos == len(self.input) {
				self.input = append(self.input, char)
			} else {
				// Insert
				self.input = append(self.input[:self.cursorPos], append([]byte{char}, self.input[self.cursorPos:]...)...)
			}
			self.cursorPos += 1
		}

		self.displayInput()

		if char == 13 { // Enter

			cleanInput := sane(string(self.input))
			fromUser <- []byte(cleanInput)
			self.input = make([]byte, 0)
			self.cursorPos = 0
		}

	}
}

// Show input so far
func (self *Terminal) displayInput() {
	self.ClearLine()

	msg := Bold("\r[" + self.Channel + "] ")
	if len(self.input) != 0 {
		width := TerminalWidth()
		inputLen := len(self.input) + len(msg)
		start := 0
		if inputLen > width {
			start = inputLen - width
		}
		visible := string(self.input[start:])

		if self.input[0] == '/' {
			// Bold IRC commands
			visible = Bold(visible)
		}

		msg += visible

		backs := len(self.input) - self.cursorPos
		if backs != 0 {
			msg += string([]byte{0x1B, '['})
			msg += strconv.Itoa(backs) + "D"
		}
	}
	self.rawWrite([]byte(msg))
}

// Width of the current terminal in columns
func TerminalWidth() int {
	sizeobj, _ := GetWinsize()
	return int(sizeobj.Col)
}

// Gets the window size using the TIOCGWINSZ ioctl() call on the tty device.
func GetWinsize() (*winsize, os.Error) {
	ws := new(winsize)

	r1, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(_TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)

	if int(r1) == -1 {
		return nil, os.NewSyscallError("GetWinsize", int(errno))
	}
	return ws, nil
}

// termios types
type cc_t byte
type speed_t uint
type tcflag_t uint

// termios constants
const (
	BRKINT = tcflag_t(0000002)
	ICRNL  = tcflag_t(0000400)
	INPCK  = tcflag_t(0000020)
	ISTRIP = tcflag_t(0000040)
	IXON   = tcflag_t(0002000)
	OPOST  = tcflag_t(0000001)
	CS8    = tcflag_t(0000060)
	ECHO   = tcflag_t(0000010)
	ICANON = tcflag_t(0000002)
	IEXTEN = tcflag_t(0100000)
	ISIG   = tcflag_t(0000001)
	VTIME  = tcflag_t(5)
	VMIN   = tcflag_t(6)
)

const NCCS = 32

type termios struct {
	c_iflag, c_oflag, c_cflag, c_lflag tcflag_t
	c_line                             cc_t
	c_cc                               [NCCS]cc_t
	c_ispeed, c_ospeed                 speed_t
}

// ioctl constants
const (
	TCGETS = 0x5401
	TCSETS = 0x5402
)

func getTermios(dst *termios) os.Error {
	r1, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(TTY_FD), uintptr(TCGETS),
		uintptr(unsafe.Pointer(dst)))

	if err := os.NewSyscallError("SYS_IOCTL", int(errno)); err != nil {
		return err
	}

	if r1 != 0 {
		return os.NewError("Error")
	}

	return nil
}

func setTermios(src *termios) os.Error {
	r1, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(TTY_FD), uintptr(TCSETS),
		uintptr(unsafe.Pointer(src)))

	if err := os.NewSyscallError("SYS_IOCTL", int(errno)); err != nil {
		return err
	}

	if r1 != 0 {
		return os.NewError("Error")
	}

	return nil
}
