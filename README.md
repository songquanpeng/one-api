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
  <a href="https://hub.docker.com/repository/docker/justsong/one-api">
    <img src="https://img.shields.io/docker/pulls/justsong/one-api?color=brightgreen" alt="docker pull">
  </a>
  <a href="https://github.com/songquanpeng/one-api/releases/latest">
    <img src="https://img.shields.io/github/downloads/songquanpeng/one-api/total?color=brightgreen&include_prereleases" alt="release">
  </a>
  <a href="https://goreportcard.com/report/github.com/songquanpeng/one-api">
    <img src="https://goreportcard.com/badge/github.com/songquanpeng/one-api" alt="GoReportCard">
  </a>
</p>

<p align="center">
  <a href="https://github.com/songquanpeng/one-api#部署">部署教程</a>
  ·
  <a href="https://github.com/songquanpeng/one-api#使用方法">使用方法</a>
  ·
  <a href="https://github.com/songquanpeng/one-api/issues">意见反馈</a>
  ·
  <a href="https://github.com/songquanpeng/one-api#截图展示">截图展示</a>
  ·
  <a href="https://openai.justsong.cn/">在线演示</a>
  ·
  <a href="https://github.com/songquanpeng/one-api#常见问题">常见问题</a>
  ·
  <a href="https://iamazing.cn/page/reward">赞赏支持</a>
</p>

> **Note**：使用 Docker 拉取的最新镜像可能是 `alpha` 版本，如果追求稳定性请手动指定版本。

> **Warning**：从 `v0.3` 版本升级到 `v0.4` 版本需要手动迁移数据库，请手动执行[数据库迁移脚本](./bin/migration_v0.3-v0.4.sql)。

## 功能
1. 支持多种 API 访问渠道，欢迎 PR 或提 issue 添加更多渠道：
   + [x] OpenAI 官方通道（支持配置代理）
   + [x] **Azure OpenAI API**
   + [x] [API2D](https://api2d.com/r/197971)
   + [x] [OhMyGPT](https://aigptx.top?aff=uFpUl2Kf)
   + [x] [AI Proxy](https://aiproxy.io/?i=OneAPI) （邀请码：`OneAPI`）
   + [x] [OpenAI-SB](https://openai-sb.com)
   + [x] [API2GPT](http://console.api2gpt.com/m/00002S)
   + [x] [CloseAI](https://console.closeai-asia.com/r/2412)
   + [x] [AI.LS](https://ai.ls)
   + [x] [OpenAI Max](https://openaimax.com)
   + [x] 自定义渠道：例如各种未收录的第三方代理服务
2. 支持通过**负载均衡**的方式访问多个渠道。
3. 支持 **stream 模式**，可以通过流式传输实现打字机效果。
4. 支持**多机部署**，[详见此处](#多机部署)。
5. 支持**令牌管理**，设置令牌的过期时间和使用次数。
6. 支持**兑换码管理**，支持批量生成和导出兑换码，可使用兑换码为账户进行充值。
7. 支持**通道管理**，批量创建通道。
8. 支持**用户分组**以及**渠道分组**，支持为不同分组设置不同的倍率。
9. 支持渠道**设置模型列表**。
10. 支持**查看额度明细**。
11. 支持发布公告，设置充值链接，设置新用户初始额度。
12. 支持丰富的**自定义**设置，
    1. 支持自定义系统名称，logo 以及页脚。
    2. 支持自定义首页和关于页面，可以选择使用 HTML & Markdown 代码进行自定义，或者使用一个单独的网页通过 iframe 嵌入。
13. 支持通过系统访问令牌访问管理 API。
14. 支持 Cloudflare Turnstile 用户校验。
15. 支持用户管理，支持**多种用户登录注册方式**：
    + 邮箱登录注册以及通过邮箱进行密码重置。
    + [GitHub 开放授权](https://github.com/settings/applications/new)。
    + 微信公众号授权（需要额外部署 [WeChat Server](https://github.com/songquanpeng/wechat-server)）。
16. 未来其他大模型开放 API 后，将第一时间支持，并将其封装成同样的 API 访问方式。

## 部署
### 基于 Docker 进行部署
部署命令：`docker run --name one-api -d --restart always -p 3000:3000 -v /home/ubuntu/data/one-api:/data justsong/one-api`

更新命令：`docker run --rm -v /var/run/docker.sock:/var/run/docker.sock containrrr/watchtower -cR`

`-p 3000:3000` 中的第一个 `3000` 是宿主机的端口，可以根据需要进行修改。

数据将会保存在宿主机的 `/home/ubuntu/data/one-api` 目录，请确保该目录存在且具有写入权限，或者更改为合适的目录。

Nginx 的参考配置：
```
server{
   server_name openai.justsong.cn;  # 请根据实际情况修改你的域名
   
   location / {
          client_max_body_size  64m;
          proxy_http_version 1.1;
          proxy_pass http://localhost:3000;  # 请根据实际情况修改你的端口
          proxy_set_header Host $host;
          proxy_set_header X-Forwarded-For $remote_addr;
          proxy_cache_bypass $http_upgrade;
          proxy_set_header Accept-Encoding gzip;
   }
}
```

之后使用 Let's Encrypt 的 certbot 配置 HTTPS：
```bash
# Ubuntu 安装 certbot：
sudo snap install --classic certbot
sudo ln -s /snap/bin/certbot /usr/bin/certbot
# 生成证书 & 修改 Nginx 配置
sudo certbot --nginx
# 根据指示进行操作
# 重启 Nginx
sudo service nginx restart
```

初始账号用户名为 `root`，密码为 `123456`。

### 手动部署
1. 从 [GitHub Releases](https://github.com/songquanpeng/one-api/releases/latest) 下载可执行文件或者从源码编译：
   ```shell
   git clone https://github.com/songquanpeng/one-api.git
   
   # 构建前端
   cd one-api/web
   npm install
   npm run build

   # 构建后端
   cd ..
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

### 多机部署
1. 所有服务器 `SESSION_SECRET` 设置一样的值。
2. 必须设置 `SQL_DSN`，使用 MySQL 数据库而非 SQLite，请自行配置主备数据库同步。
3. 所有从服务器必须设置 `SYNC_FREQUENCY`，以定期从数据库同步配置。
4. 从服务器可以选择设置 `FRONTEND_BASE_URL`，以重定向页面请求到主服务器。

环境变量的具体使用方法详见[此处](#环境变量)。

### 宝塔部署教程

详见[#175](https://github.com/songquanpeng/one-api/issues/175)。

如果部署后访问出现空白页面，详见[#97](https://github.com/songquanpeng/one-api/issues/97)。

### 部署第三方服务配合 One API 使用
> 欢迎 PR 添加更多示例。

#### ChatGPT Next Web
项目主页：https://github.com/Yidadaa/ChatGPT-Next-Web

```bash
docker run --name chat-next-web -d -p 3001:3000 -e BASE_URL=https://openai.justsong.cn yidadaa/chatgpt-next-web
```

注意修改端口号和 `BASE_URL`。

#### ChatGPT Web
项目主页：https://github.com/Chanzhaoyu/chatgpt-web

```bash
docker run --name chatgpt-web -d -p 3002:3002 -e OPENAI_API_BASE_URL=https://openai.justsong.cn -e OPENAI_API_KEY=sk-xxx chenzhaoyu94/chatgpt-web
```

注意修改端口号、`OPENAI_API_BASE_URL` 和 `OPENAI_API_KEY`。

### 部署到第三方平台
<details>
<summary><strong>部署到 Zeabur</strong></summary>
<div>

> Zeabur 的服务器在国外，自动解决了网络的问题，同时免费的额度也足够个人使用。

1. 首先 fork 一份代码。
2. 进入 [Zeabur](https://zeabur.com/)，登录，进入控制台。
3. 新建一个 Project，在 Service -> Add Service 选择 Marketplace，选择 MySQL，并记下连接参数（用户名、密码、地址、端口）。
4. 复制链接参数，运行 ```create database `one-api` ``` 创建数据库。
5. 然后在 Service -> Add Service，选择 Git（第一次使用需要先授权），选择你 fork 的仓库。
6. Deploy 会自动开始，先取消。进入下方 Variable，添加一个 `PORT`，值为 `3000`，再添加一个 `SQL_DSN`，值为 `<username>:<password>@tcp(<addr>:<port>)/one-api` ，然后保存。 注意如果不填写 `SQL_DSN`，数据将无法持久化，重新部署后数据会丢失。
7. 选择 Redeploy。
8. 进入下方 Domains，选择一个合适的域名前缀，如 "my-one-api"，最终域名为 "my-one-api.zeabur.app"，也可以 CNAME 自己的域名。
9. 等待部署完成，点击生成的域名进入 One API。

</div>
</details>

## 配置
系统本身开箱即用。

你可以通过设置环境变量或者命令行参数进行配置。

等到系统启动后，使用 `root` 用户登录系统并做进一步的配置。

## 使用方法
在`渠道`页面中添加你的 API Key，之后在`令牌`页面中新增访问令牌。

之后就可以使用你的令牌访问 One API 了，使用方式与 [OpenAI API](https://platform.openai.com/docs/api-reference/introduction) 一致。

你需要在各种用到 OpenAI API 的地方设置 API Base 为你的 One API 的部署地址，例如：`https://openai.justsong.cn`，API Key 则为你在 One API 中生成的令牌。

注意，具体的 API Base 的格式取决于你所使用的客户端。

```mermaid
graph LR
    A(用户)
    A --->|请求| B(One API)
    B -->|中继请求| C(OpenAI)
    B -->|中继请求| D(Azure)
    B -->|中继请求| E(其他下游渠道)
```

可以通过在令牌后面添加渠道 ID 的方式指定使用哪一个渠道处理本次请求，例如：`Authorization: Bearer ONE_API_KEY-CHANNEL_ID`。
注意，需要是管理员用户创建的令牌才能指定渠道 ID。

不加的话将会使用负载均衡的方式使用多个渠道。

### 环境变量
1. `REDIS_CONN_STRING`：设置之后将使用 Redis 作为请求频率限制的存储，而非使用内存存储。
   + 例子：`REDIS_CONN_STRING=redis://default:redispw@localhost:49153`
2. `SESSION_SECRET`：设置之后将使用固定的会话密钥，这样系统重新启动后已登录用户的 cookie 将依旧有效。
   + 例子：`SESSION_SECRET=random_string`
3. `SQL_DSN`：设置之后将使用指定数据库而非 SQLite，请使用 MySQL 8.0 版本。
   + 例子：`SQL_DSN=root:123456@tcp(localhost:3306)/one-api`
4. `FRONTEND_BASE_URL`：设置之后将使用指定的前端地址，而非后端地址。
   + 例子：`FRONTEND_BASE_URL=https://openai.justsong.cn`
5. `SYNC_FREQUENCY`：设置之后将定期与数据库同步配置，单位为秒，未设置则不进行同步。
   + 例子：`SYNC_FREQUENCY=60`

### 命令行参数
1. `--port <port_number>`: 指定服务器监听的端口号，默认为 `3000`。
   + 例子：`--port 3000`
2. `--log-dir <log_dir>`: 指定日志文件夹，如果没有设置，日志将不会被保存。
   + 例子：`--log-dir ./logs`
3. `--version`: 打印系统版本号并退出。
4. `--help`: 查看命令的使用帮助和参数说明。

## 演示
### 在线演示
注意，该演示站不提供对外服务：
https://openai.justsong.cn

### 截图展示
![channel](https://user-images.githubusercontent.com/39998050/233837954-ae6683aa-5c4f-429f-a949-6645a83c9490.png)
![token](https://user-images.githubusercontent.com/39998050/233837971-dab488b7-6d96-43af-b640-a168e8d1c9bf.png)

## 常见问题
1. 账户额度足够为什么提示额度不足？
   + 请检查你的令牌额度是否足够，这个和账户额度是分开的。
   + 令牌额度仅供用户设置最大使用量，用户可自由设置。
2. 提示无可用渠道？
   + 请检查的用户分组和渠道分组设置。
   + 以及渠道的模型设置。
3. 渠道测试报错：`invalid character '<' looking for beginning of value`
   + 这是因为返回值不是合法的 JSON，而是一个 HTML 页面。
   + 大概率是你的部署站的 IP 或代理的节点被 CloudFlare 封禁了。

## 注意
本项目为开源项目，请在遵循 OpenAI 的[使用条款](https://openai.com/policies/terms-of-use)以及法律法规的情况下使用，不得用于非法用途。

本项目使用 MIT 协议进行开源，请以某种方式保留 One API 的版权信息。

依据 MIT 协议，使用者需自行承担使用本项目的风险与责任，本开源项目开发者与此无关。