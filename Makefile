prefix = /usr/local
bindir = $(prefix)/bin

SRCS = main.go

VERSION = 0.1.1

.PHONY: all
all: ijq

ijq: $(SRCS)
	go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $@

.PHONY: install
install: ijq
	install -d $(bindir)
	install -m 0755 $< $(bindir)

.PHONY: uninstall
uninstall:
	rm $(bindir)/ijq

.PHONY: clean
clean:
	rm ijq
