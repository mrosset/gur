include $(GOROOT)/src/Make.inc

TARG=gur
GOFILES=gur.go aur.go
GOFMT=gofmt -s -l -w

include $(GOROOT)/src/Make.cmd

format:
	${GOFMT} .

test: format all
	time ./${TARG} -d gobuild-hg
