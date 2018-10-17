VERSION=0.0.1

.PHONY: all run build clean proto

all: clean proto build

help: 
	@echo 
	@echo "toy"
	@echo 

proto:
	protoc --go_out=. ./protocol/*.proto
	go run script/genProtoId.go protocol

run:
	go run main.go

build:
	go build

clean:
	go clean
