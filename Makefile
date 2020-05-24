PREFIX := /usr/local
BINDIR := $(PREFIX)/bin
SRCS := main.go

.PHONY: all
all: ijq

ijq: $(SRCS)
	go build -ldflags="-s -w" -o $@

.PHONY: install
install: ijq
	mkdir -p $(BINDIR)
	install -m 0755 $< $(BINDIR)

.PHONY: clean
clean:
	rm ijq
