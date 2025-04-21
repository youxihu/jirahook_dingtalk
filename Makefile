.PHONY: build
# build
build:
	rm -rf ./bin
	mkdir -p bin/ && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o ./bin/ ./...
	upx  ./bin/*

docker:
	docker build -t "192.168.2.254:54800/tools-jira-hook:v0.0.1" .
