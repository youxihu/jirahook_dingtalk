version := $(shell cat VERSION)

.PHONY: build
# build
build:
	rm -rf ./bin
	mkdir -p bin/ && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o ./bin/ ./...
	upx  ./bin/*

docker:
	docker build -t "192.168.2.254:54800/tools/jira-hook:$(version)" .

run:
	docker run -di \
            --name jira_hook \
            -p 4165:4165 \
            -v /usr/local/secret/dingtalk/secret.yaml:/app-acc/dingtalk/secret.yaml \
            -v /usr/local/secret/phonenumber/secret.yaml:/app-acc/phonenumber/secret.yaml \
            192.168.2.254:54800/tools/jira-hook:$(version)