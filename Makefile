.PHONY: all test build-binary clean

GO ?= go
GO_CMD := CGO_ENABLED=0 $(GO)

all: test build-binary

test:
	$(GO_CMD) test -cover ./...

build-binary:
	$(GO_CMD) build -tags netgo -ldflags "-w" -o bukky .

clean:
	rm -f bukky
