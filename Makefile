GOPATH	= $(CURDIR)
BINDIR	= $(CURDIR)/bin

PROGRAMS = check_redfish

build:
	env GOPATH=$(GOPATH) go install $(PROGRAMS)

destdirs:
	mkdir -p -m 0755 $(DESTDIR)/usr/bin

strip: build
	strip --strip-all $(BINDIR)/check_redfish

install: strip destdirs install-bin

install-bin:
	install -m 0755 $(BINDIR)/check_redfish $(DESTDIR)/usr/bin

clean:
	/bin/rm -f bin/check_redfish

uninstall:
	/bin/rm -f $(DESTDIR)/usr/bin

all: build strip install

