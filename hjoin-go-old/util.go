package main

import (
	"strings"
)

const (
	BLACK     = "30"
	GREY      = "1;30"
	LIGHTGREY = "37"
	WHITE     = "1;37"

	RED      = "31"
	LIGHTRED = "1;31"

	GREEN      = "32"
	LIGHTGREEN = "1;32"

	ORANGE = "33"
	YELLOW = "1;33"

	BLUE      = "34"
	LIGHTBLUE = "1;34"

	PURPLE = "35"
	PINK   = "1;35"

	CYAN      = "36"
	LIGHTCYAN = "1;36"
)

var UserColors = []string{RED, GREEN, ORANGE, LIGHTBLUE, YELLOW, PURPLE, CYAN, LIGHTRED, LIGHTGREEN, PINK, LIGHTCYAN}

/* Trims a string to not include junk such as:
- the null bytes after a character return
- \n and \r
- whitespace
- Ascii char \001, which is the extended data delimiter,
  used for example in a /me command before 'ACTION'.
  See http://www.irchelp.org/irchelp/rfc/ctcpspec.html
*/
func sane(data string) string {
	parts := strings.SplitN(data, "\n", 2)
	return strings.Trim(parts[0], " \n\r\001")
}

func Bold(str string) string {
	return "\033[1m" + str + "\033[0m"
}

func Underline(str string) string {
	return "\033[4m" + str + "\033[0m"
}

// color is one of the constants in this file
func Color(color string, str string) string {
	return "\033[" + color + "m" + str + "\033[0m"
}

// Blue on white
func highlight(str string) string {
	return "\033[34;47m" + str + "\033[0m"
}

// Pad a string left
func Lpad(chars int, str string) string {
	for len(str) < chars {
		str = " " + str
	}
	return str
}
