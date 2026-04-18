.PHONY: pb-generate generate test format

pb-generate:
	buf generate https://github.com/heptaliane/katarive-proto.git --path api

generate:
	go generate ./...

test:
	go test ./...

format:
	go fmt ./...

run:
	go run ./...
