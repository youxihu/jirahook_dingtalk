version := $(shell cat VERSION)

.PHONY: build docker docker_run docker_push
# build
build:
	rm -rf ./bin
	mkdir -p bin/ && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build  -o ./bin/ ./...
	upx  ./bin/*

docker:
	docker build -t "192.168.2.254:54800/tools/jira-hook:$(version)" .

docker_run:
	docker run -di \
            --name jira_hook \
            -p 4165:4165 \
            -v /home/youxihu/secret/dingtalk/secret.yaml:/app-acc/dingtalk/secret.yaml \
            -v /home/youxihu/secret/phonenumber/secret.yaml:/app-acc/phonenumber/secret.yaml \
            192.168.2.254:54800/tools/jira-hook:$(version)


docker_push:
	docker push 192.168.2.254:54800/tools/jira-hook:$(version)