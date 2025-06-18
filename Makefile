version := $(shell cat VERSION)

.PHONY: build docker docker_run docker_push
#run
run:
	go run cmd/jira_hook/main.go

#test
test:
	cd internal/handler && go test -v
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
            -v /home/youxihu/secret/jira_hook/dingcfg.yaml:/app-acc/configs/dingcfg.yaml \
            -v /home/youxihu/secret/jira_hook/phonenumb.yaml:/app-acc/configs/phonenumb.yaml \
            -v /home/youxihu/secret/jira_hook/redis.yaml:/app-acc/configs/redis.yaml \
            -v /home/youxihu/secret/jira_hook/mysql.yaml:/app-acc/configs/mysql.yaml \
            192.168.2.254:54800/tools/jira-hook:$(version)

docker_push:
	docker push 192.168.2.254:54800/tools/jira-hook:$(version)