.PHONY: run
run:
	go run cmd/ibot/main.go -config ./config.json

.PHONY: build
build:
	go build -v -o bin/ibot cmd/ibot/main.go

