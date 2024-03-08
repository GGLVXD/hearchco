run:
	air
run-cli:
	go run ./src --cli

debug:
	air -- -v
debug-cli:
	go run ./srv -v --cli

trace:
	air -- -vv
trace-cli:
	go run ./src -vv --cli

install:
	go get ./...
	go install github.com/dmarkham/enumer@latest
	go generate ./...
	go install github.com/cosmtrek/air@latest

build:
	go build ./...

test:
	sh ./scripts/test.sh
test-engines:
	sh ./scripts/test-engines.sh
test-redis:
	sh ./scripts/test-redis.sh
test-redis-podman:
	sh ./scripts/test-redis-podman.sh
test-redis-docker:
	sh ./scripts/test-redis-docker.sh
test-all: test test-redis test-engines
test-all-podman: test test-redis-podman test-engines
test-all-docker: test test-redis-docker test-engines

update:
	go get -u ./...
	go mod tidy

lint:
	golangci-lint run
