include $(GOROOT)/src/Make.inc

TARG=gur
GOFILES=gur.go aur.go goarchive.go pacman.go
GOFMT=gofmt -s -l -w
CC=gccgo-4.7
CFLAGS=-g -O2 -pipe

include $(GOROOT)/src/Make.cmd

CLEANFILES+=./tmp/*


test: clean all
	#gotest
	cd ./tmp;../${TARG} git
	#cd ./tmp;../${TARG} -d yi

format:
	${GOFMT} .

gur.o:
	${CC} ${CFLAGS} -c ${GOFILES} -o gur.o

gccgo: gur.o
	${CC} ${CFLAGS} gur.o -o ${TARG}

install-gccgo: gccgo
	cp -f ${TARG} ${GOROOT}/bin
