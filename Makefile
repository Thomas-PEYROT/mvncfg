.PHONY: build install test vet clean fmt

BINARY_NAME := mvncfg
CMD_DIR := ./cmd/mvncfg
INSTALL_DIR := $(HOME)/.local/bin

build:
	go build -o $(BINARY_NAME) $(CMD_DIR)

install: build
	mkdir -p $(INSTALL_DIR)
	cp -f $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY_NAME)

fmt:
	gofmt -w .
