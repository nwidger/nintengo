# This Makefile exists to help make a Mac OS X app bundle.  If you
# want to build the regular nintengo binary, just use 'go build'.

all: nintengo Nintengo.app

Nintengo.app:
	cp nintengo Nintengo.app/Contents/MacOS

nintengo:
	go generate ./...
	go build

.PHONY: nintengo Nintengo.app
