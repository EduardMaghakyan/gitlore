.PHONY: build install test lint clean

build:
	go build -o gitlore .

install: build
	mv gitlore /usr/local/bin/

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f gitlore
