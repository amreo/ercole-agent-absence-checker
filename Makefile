# Simple Makefile for ercole agent

DESTDIR=build

all: ercole-agent-absence-checker

default: ercole-agent-absence-checker

clean:
	rm -rf ercole-agent-absence-checker build ercole-agent-absence-checker.exe *.exe

ercole-agent:
	GO111MODULE=on CGO_ENABLED=0 go build -o ercole-agent-absence-checker -a -x

windows:
	GOOS=windows GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 go build -o ercole-agent-absence-checker.exe -a -x

nsis: windows 
	makensis package/win/installer.nsi

install: all install-fetchers install-bin install-bin install-config install-scripts

install-fetchers:
	install -d $(DESTDIR)/fetch
	cp -rp fetch/* $(DESTDIR)/fetch
	rm $(DESTDIR)/fetch/*.ps1

install-bin:
	install -m 755 ercole-agent-absence-checker $(DESTDIR)/ercole-agent-absence-checker

install-scripts:
	install -d $(DESTDIR)/sql
	install -m 644 sql/*.sql $(DESTDIR)/sql

install-config:
	install -m 644 config.json $(DESTDIR)/config.json
