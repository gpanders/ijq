prefix = /usr/local
bindir = $(prefix)/bin
mandir = $(prefix)/share/man

SRCS = main.go history.go

VERSION = 0.4.1

.PHONY: all
all: ijq docs

.PHONY: docs
docs: ijq.1

ijq: $(SRCS)
	go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $@

%.1: %.1.scd
	scdoc < $< > $@

.PHONY: test
test:
	go test -v -coverprofile=./cover.out .

.PHONY: viewcover
viewcover:
	go tool cover -html=./cover.out

.PHONY: install
install: ijq ijq.1
	install -d $(DESTDIR)$(bindir) $(DESTDIR)$(mandir)/man1
	install -m 0755 ijq $(DESTDIR)$(bindir)
	install -m 0644 ijq.1 $(DESTDIR)$(mandir)/man1

.PHONY: uninstall
uninstall:
	rm $(DESTDIR)$(bindir)/ijq $(DESTDIR)$(mandir)/man1/ijq.1

.PHONY: clean
clean:
	rm -f ijq ijq.1
