all: clean build

build:
	find . -name "*.proto" -type f -print0 | xargs protoc --gogoslick_out=.
	go build firempq

install:
	go install firempq

proto:
	build proto

clean:
	go clean ./...

tests:
	go test ./...

vet:
	go vet ./...
	go tool vet --shadow .
