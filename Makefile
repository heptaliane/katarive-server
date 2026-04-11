pb-generate:
	buf generate https://github.com/heptaliane/katarive-proto.git --path api

test:
	go test ./...

generate:
	go generate ./...
