
include ${GOROOT}/src/Make.inc

all:
	gomake -C hatcogd
	gomake -C hjoin

clean:
	gomake -C hatcogd clean
	gomake -C hjoin clean

format:
	find . -type f -name "*.go" -exec gofmt -w {} \;

