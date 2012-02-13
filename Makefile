GOFMT=gofmt -s -l -w
CC=gccgo-4.7
CFLAGS=-g -O2 -pipe

CLEANFILES+=./tmp/*

gur:
	go build

test: 
	go test

format:
	${GOFMT} .

gur.o:
	${CC} ${CFLAGS} -c ${GOFILES} -o gur.o

gccgo: gur.o
	${CC} ${CFLAGS} gur.o -o ${TARG}

install-gccgo: gccgo
	cp -f ${TARG} ${GOROOT}/bin
