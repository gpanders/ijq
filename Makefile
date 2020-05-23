SRCS := main.go

.PHONY: all
all: ijq

ijq: $(SRCS)
	go build

.PHONY: clean
clean:
	rm ijq
