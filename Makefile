.PHONY: run
run:
	go run main.go -config ./config.json

.PHONY: build
build:
	go build -v -o ibot main.go

