.PHONY: build install clean test

build:
	go build -o bin/ocsession cmd/ocsession/main.go

install:
	go build -o /usr/local/bin/ocsession cmd/ocsession/main.go

clean:
	rm -rf bin/
	go clean

test:
	go test -v ./test/unit/...

run:
	go run cmd/ocsession/main.go