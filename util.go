package main

import (
    "strings"
)

/* Trims a string to not include junk such as:
 - the null bytes after a character return
 - \n and \r
 - whitespace
*/
func sane(data string) string {
    parts := strings.SplitN(data, "\n", 2)
    return strings.Trim(parts[0], " \n\r")
}

