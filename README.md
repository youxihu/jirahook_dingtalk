# jirahook_dingtalk

`jirahook_dingtalk` 是一个用于将 Jira Webhook 事件转发至钉钉群机器人的服务组件。该服务可实现 Jira 事件的实时通知推送，提升团队协作效率与信息同步能力。

---

## 功能概述

- **Webhook 接收支持**：可监听并接收来自 Jira 的 Webhook 事件。
- **钉钉消息通知**：对事件内容进行结构化处理后发送至钉钉群机器人。
- **容器化部署支持**：内置 Docker 构建能力，便于快速部署与迁移。
- **高度可配置性**：通过外部配置文件灵活管理 Jira 与钉钉的集成参数。

---

## 快速开始

### 1. 构建可执行文件

使用以下命令编译项目：

```bash
make build
```

构建过程将清理旧的构建产物，并在 `./bin` 目录下生成名为 `jira_hook` 的可执行文件。如需自定义命名或构建逻辑，可修改 `Makefile`。

---

### 2. 构建 Docker 镜像

```bash
make docker
```

该命令基于根目录下的 `Dockerfile` 构建镜像，默认标签为：
`YourRepoAddr/tools-jira-hook:v1.x.x`
可根据实际仓库地址进行调整。

---

### 3. 准备配置文件

运行前需准备以下两份 YAML 配置文件，并在容器中正确挂载。

#### `phoneNumber.yaml`

```yaml
游西湖: 1981546502
Jira管理员: 6546121
杰尼龟: 1564512312
约翰逊: 1651231321
```

> 配置说明：键为 Jira 中的用户名称（name 字段），值为钉钉群成员对应手机号。用于在消息推送中精确 @ 相关成员。

---

#### `dingtalk_webhookUrl.yaml`

```yaml
DINGTALK_ACCESS_TOKEN: "https://oapi.dingtalk.com/robot/send?access_token=122e74d4282557e981262d9aa23c5"
DINGTALK_SECRET: "SEC0aed4136ff2dasfsdafsdfbaksdjsbd365e81fb367d71fc6"  # 可选，若启用签名校验需配置
```

---

### 4. 启动服务容器

以下示例为标准的 Docker 启动命令：

```bash
docker run -di --name jira_hook \
  -p 4165:4165 \
  -v /usr/local/secret/phoneNumber.yaml:/app-acc/phonenumber/secret.yaml \
  -v /usr/local/secret/dingtalk_webhookUrl.yaml:/app-acc/dingtalk/secret.yaml \
  192.168.2.254:54800/tools-jira-hook:v0.0.1
```

---

至此，`jirahook_dingtalk` 服务已成功启动并可实时接收 Jira 推送的事件信息，自动转发至指定的钉钉群。可进一步结合团队需求，扩展事件处理逻辑与消息格式。
