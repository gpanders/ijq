PREFIX := /usr/local
BINDIR := $(PREFIX)/bin
SRCS := main.go

VERSION := 0.1.0

.PHONY: all
all: ijq

ijq: $(SRCS)
	go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $@

.PHONY: install
install: ijq
	mkdir -p $(BINDIR)
	install -m 0755 $< $(BINDIR)

.PHONY: clean
clean:
	rm ijq
