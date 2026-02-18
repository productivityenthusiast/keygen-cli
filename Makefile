VERSION := 0.1.0
BINARY := keygen
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build clean test lint install run

build:
	go build -buildvcs=false -ldflags "$(LDFLAGS)" -o $(BINARY).exe .

clean:
	rm -f $(BINARY) $(BINARY).exe

test:
	go test -race -cover ./...

lint:
	golangci-lint run

install: build
	cp $(BINARY).exe c:/data/dev/$(BINARY).exe

run:
	go run -ldflags "$(LDFLAGS)" . $(ARGS)

deploy: build install
	@echo "Deployed $(BINARY) v$(VERSION)"
