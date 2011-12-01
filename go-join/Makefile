# Copyright 2009 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include ${GOROOT}/src/Make.inc

TARG=go-join
GOFILES=\
	main.go\
	user_manager.go\
	internal-connection.go\
	line.go\
	util.go\
	term_input.go\
	term.go

include ${GOROOT}/src/Make.cmd

format:
	find . -type f -name "*.go" -exec gofmt -w {} \;
