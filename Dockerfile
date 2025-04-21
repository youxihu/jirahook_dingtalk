# base image
FROM 192.168.2.254:54800/alpine:latest

# 创建目标文件夹
RUN mkdir -p /app-acc/configs

# 设置固定的项目路径
ENV WORKDIR /app-acc

# 设置时区为中国上海
RUN echo -e  "http://mirrors.aliyun.com/alpine/v3.4/main\nhttp://mirrors.aliyun.com/alpine/v3.4/community" >  /etc/apk/repositories \
&& apk update && apk add tzdata \
&& cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
&& echo "Shanghai/Asia" > /etc/timezone \
&& apk del tzdata

COPY ./bin/jira_hook   $WORKDIR/jira_hook

# run shell cmd
RUN chmod +x $WORKDIR/jira_hook

# work space
WORKDIR $WORKDIR

# open port
EXPOSE 4165

# start
CMD ["./jira_hook"]
