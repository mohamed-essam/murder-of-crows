export GOPATH = $(shell pwd)

install:
	go get gopkg.in/redis.v5

build: install
	go build murder.go

run: install
	go run murder.go
