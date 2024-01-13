.PHONY: run
run:
	go run cmd/main.go -config ./config.json

.PHONY: build
build:
	go build -v -o bin/ibot cmd/main.go

