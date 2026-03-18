build:
	go build -o bin/hookscope ./...

install:
	go install ./...

.DEFAULT_GOAL := build
