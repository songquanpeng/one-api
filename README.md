<p align="center">
  <a href="https://github.com/songquanpeng/one-api"><img src="https://raw.githubusercontent.com/songquanpeng/one-api/main/web/public/logo.png" width="150" height="150" alt="one-api logo"></a>
</p>

<div align="center">

# One API

_✨ All in one 的 OpenAI 接口，整合各种 API 访问方式，开箱即用✨_

</div>

<p align="center">
  <a href="https://raw.githubusercontent.com/songquanpeng/one-api/main/LICENSE">
    <img src="https://img.shields.io/github/license/songquanpeng/one-api?color=brightgreen" alt="license">
  </a>
  <a href="https://github.com/songquanpeng/one-api/releases/latest">
    <img src="https://img.shields.io/github/v/release/songquanpeng/one-api?color=brightgreen&include_prereleases" alt="release">
  </a>
  <a href="https://github.com/songquanpeng/one-api/releases/latest">
    <img src="https://img.shields.io/github/downloads/songquanpeng/one-api/total?color=brightgreen&include_prereleases" alt="release">
  </a>
  <a href="https://goreportcard.com/report/github.com/songquanpeng/one-api">
    <img src="https://goreportcard.com/badge/github.com/songquanpeng/one-api" alt="GoReportCard">
  </a>
</p>

<p align="center">
  <a href="https://github.com/songquanpeng/one-api/releases">程序下载</a>
  ·
  <a href="https://github.com/songquanpeng/one-api#部署">部署教程</a>
  ·
  <a href="https://github.com/songquanpeng/one-api/issues">意见反馈</a>
  ·
  <a href="https://github.com/songquanpeng/one-api#截图展示">截图展示</a>
  ·
  <a href="https://openai.justsong.cn/">在线演示</a>
</p>

## 功能
1. 支持多种 API 访问渠道，欢迎 PR 或提 issue 添加更多渠道：
   + [x] One API 服务端中继
   + [x] [API2D](https://api2d.com/r/197971)
   + [ ] Azure OpenAI API
   + [x] [CloseAI](https://console.openai-asia.com)
   + [x] [OpenAI-SB](https://openai-sb.com)
   + [x] [OpenAI Max](https://openaimax.com)
   + [x] [OhMyGPT](https://www.ohmygpt.com)
   + [x] 自定义渠道
2. 支持通过负载均衡的方式访问多个渠道。
3. 支持单个访问渠道设置多个 API Key，利用起来你的多个 API Key。
4. 支持 HTTP SSE。
5. 多种用户登录注册方式：
   + 邮箱登录注册以及通过邮箱进行密码重置。
   + [GitHub 开放授权](https://github.com/settings/applications/new)。
   + 微信公众号授权（需要额外部署 [WeChat Server](https://github.com/songquanpeng/wechat-server)）。
6. 支持用户管理。

## 部署
### 基于 Docker 进行部署
执行：`docker run -d --restart always -p 3000:3000 -v /home/ubuntu/data/one-api:/data -v /etc/ssl/certs:/etc/ssl/certs:ro justsong/one-api`

数据将会保存在宿主机的 `/home/ubuntu/data/one-api` 目录。

### 手动部署
1. 从 [GitHub Releases](https://github.com/songquanpeng/one-api/releases/latest) 下载可执行文件或者从源码编译：
   ```shell
   git clone https://github.com/songquanpeng/one-api.git
   go mod download
   go build -ldflags "-s -w" -o one-api
   ````
2. 运行：
   ```shell
   chmod u+x one-api
   ./one-api --port 3000 --log-dir ./logs
   ```
3. 访问 [http://localhost:3000/](http://localhost:3000/) 并登录。初始账号用户名为 `root`，密码为 `123456`。

更加详细的部署教程[参见此处](https://iamazing.cn/page/how-to-deploy-a-website)。

## 配置
系统本身开箱即用。

你可以通过设置环境变量或者命令行参数进行配置。

等到系统启动后，使用 `root` 用户登录系统并做进一步的配置。

## 使用方式
在`渠道`页面中添加你的 API Key，之后在`令牌`页面中新增一个访问令牌。

之后就可以使用你的令牌访问 One API 了，使用方式与 [OpenAI API](https://platform.openai.com/docs/api-reference/introduction) 一致。

可以通过在令牌后面添加渠道 ID 的方式指定使用哪一个渠道处理本次请求，例如：`Authorization: Bearer ONE_API_KEY-CHANNEL_ID`。

不加的话将会使用负载均衡的方式使用多个渠道。

### 环境变量
1. `REDIS_CONN_STRING`：设置之后将使用 Redis 作为请求频率限制的存储，而非使用内存存储。
   + 例子：`REDIS_CONN_STRING=redis://default:redispw@localhost:49153`
2. `SESSION_SECRET`：设置之后将使用固定的会话密钥，这样系统重新启动后已登录用户的 cookie 将依旧有效。
   + 例子：`SESSION_SECRET=random_string`
3. `SQL_DSN`：设置之后将使用指定数据库而非 SQLite。
   + 例子：`SQL_DSN=root:123456@tcp(localhost:3306)/one-api`

### 命令行参数
1. `--port <port_number>`: 指定服务器监听的端口号，默认为 `3000`。
   + 例子：`--port 3000`
2. `--log-dir <log_dir>`: 指定日志文件夹，如果没有设置，日志将不会被保存。
   + 例子：`--log-dir ./logs`
3. `--version`: 打印系统版本号并退出。

## 演示
### 在线演示
注意，该演示站不提供对外服务：
https://openai.justsong.cn

### 截图展示
![channel](https://user-images.githubusercontent.com/39998050/233837954-ae6683aa-5c4f-429f-a949-6645a83c9490.png)
![token](https://user-images.githubusercontent.com/39998050/233837971-dab488b7-6d96-43af-b640-a168e8d1c9bf.png)
