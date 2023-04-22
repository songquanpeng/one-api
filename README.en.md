<p align="right">
    <a href="./README.md">中文</a> | <strong>English</strong>
</p>

<p align="center">
  <a href="https://github.com/songquanpeng/one-api"><img src="https://raw.githubusercontent.com/songquanpeng/one-api/main/web/public/logo.png" width="150" height="150" alt="one-api logo"></a>
</p>

<div align="center">

# Gin Template

_✨ Template for Gin & React projects ✨_

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
  <a href="https://goreportcard.com/report/github.com/songquanpeng/go-file">
    <img src="https://goreportcard.com/badge/github.com/songquanpeng/one-api" alt="GoReportCard">
  </a>
</p>

<p align="center">
  <a href="https://github.com/songquanpeng/one-api/releases">Download</a>
  ·
  <a href="https://github.com/songquanpeng/one-api/blob/main/README.en.md#deployment">Tutorial</a>
  ·
  <a href="https://github.com/songquanpeng/one-api/issues">Feedback</a>
  ·
  <a href="https://one-api.vercel.app/">Demo</a>
</p>

## Features
+ [x] Built-in user management.
+ [x] Built-in file management.
+ [x] [GitHub OAuth](https://github.com/settings/applications/new).
+ [x] WeChat official account authorization (need [wechat-server](https://github.com/songquanpeng/wechat-server)).
+ [x] Email verification & password reset.
+ [x] Request rate limiting.
+ [x] Static files caching.
+ [x] Mobile friendly UI.
+ [x] Token based authorization.
+ [x] Use GitHub Actions to build releases & Docker images.
+ [x] Cloudflare Turnstile user validation.

## Deployment
### Manual deployment
1. Download built binary from [GitHub Releases](https://github.com/songquanpeng/one-api/releases/latest) or build from source:
   ```shell
   git clone https://github.com/songquanpeng/one-api.git
   go mod download
   go build -ldflags "-s -w" -o one-api
   ````
2. Run it:
   ```shell
   chmod u+x one-api
   ./one-api --port 3000 --log-dir ./logs
   ```
3. Visit [http://localhost:3000/](http://localhost:3000/) and login. The username for the initial account is `root` and the password is `123456`.

### Deploy with Docker
Execute: `docker run -d --restart always -p 3000:3000 -v /home/ubuntu/data/one-api:/data -v /etc/ssl/certs:/etc/ssl/certs:ro justsong/one-api`

Data will be saved in `/home/ubuntu/data/one-api`.

## Configurations
The system works out of the box.

You can configure the system by set environment variables or specify command line arguments.

After the system starts, use `root` user to log in to the system and do further configuration.

### Environment Variables
1. `REDIS_CONN_STRING`: when set, Redis will be used as the storage for request rate limitation instead of memory storage.
   + Example: `REDIS_CONN_STRING=redis://default:redispw@localhost:49153`
2. `SESSION_SECRET`: when set, a fixed session key will be used so that the logged-in users' cookie remains valid across system reboots.
   + Example: `SESSION_SECRET=random_string`
3. `SQL_DSN`: when set, the target SQL database will be used instead of SQLite.
   + Example: `SQL_DSN=root:123456@tcp(localhost:3306)/one-api`

### Command line Arguments
1. `--port <port_number>`: specify the port number, the default value is `3000`.
   + Example: `--port 3000`
2. `--log-dir <log_dir>`: specify the log dir, if not set, the log won't be saved.
   + Example: `--log-dir ./logs`
3. `--version`: print the version and exit.