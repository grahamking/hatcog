# Copyright 2009 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include ${GOROOT}/src/Make.inc

TARG=goirc
GOFILES=\
	main.go\
	connection.go\
	line.go\
	user_side.go\
	util.go\
	term.go

include ${GOROOT}/src/Make.cmd

