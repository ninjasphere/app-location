all:
	scripts/build.sh

qa: vet lint test

lint:
	go get github.com/golang/lint/golint
	$(GOPATH)/bin/golint ./...

clean:
	rm -f bin/* || true
	rm -rf .gopath || true

test:
	go test -v ./...

vet:
	go vet ./...

here: build qa

build:
	go build -o bin/app-location

.PHONY: all qa lint clean test vet here
