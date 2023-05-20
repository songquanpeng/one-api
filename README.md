<p align="center">
   <a href="https://github.com/songquanpeng/one-api"><img src="https://raw.githubusercontent.com/songquanpeng/one-api/main/web/public/logo.png " width="150" height="150" alt="one-api logo"></a>
</p>

<div align="center">

# One API

_✨All in one's OpenAI interface, integrating various API access methods, out of the box✨_

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
   <a href="https://github.com/songquanpeng/one-api/releases">Program Download</a>
   ·
   <a href="https://github.com/songquanpeng/one-api#deployment">Deployment Tutorial</a>
   ·
   <a href="https://github.com/songquanpeng/one-api/issues">Feedback</a>
   ·
   <a href="https://github.com/songquanpeng/one-api#Screenshot display">Screenshot display</a>
   ·
   <a href="https://openai.justsong.cn/">Online Demo</a>
   ·
   <a href="https://github.com/songquanpeng/one-api#FAQ">FAQ</a>
</p>

> **Warning**: Upgrading from `v0.2` to `v0.3` requires manual database migration, please manually execute [database migration script](./bin/migration_v0.2-v0.3.sql) .


## Function
1. Multiple API access channels are supported. PRs or issues are welcome to add more channels:
    + [x] OpenAI official channel
    + [x] **Azure OpenAI API**
    + [x] [API2D](https://api2d.com/r/197971)
    + [x] [OhMyGPT](https://aigptx.top?aff=uFpUl2Kf)
    + [x] [AI Proxy](https://aiproxy.io/?i=OneAPI)
   + [x] [AI.LS](https://ai.ls)
    + [x] [OpenAI Max](https://openaimax.com)
    + [x] [OpenAI-SB](https://openai-sb.com)
    + [x] [CloseAI](https://console.openai-asia.com)
    + [x] Custom channels: e.g. using a self-built OpenAI agent
2. Support access to multiple channels through **load balancing**.
3. Support **stream mode**, you can achieve typewriter effect through streaming.
4. Support **multi-machine deployment**, [see here for details](#multi-machine deployment).
5. Support **token management**, set the expiration time and usage times of the token.
6. Supports **redemption code management**, supports batch generation and export of redemption codes, and can use redemption codes to recharge accounts.
7. Support **channel management**, create channels in batches.
8. Support for publishing announcements, setting recharge links, and setting initial quotas for new users.
9. Support rich **custom** settings,
    1. Support custom system name, logo and footer.
    2. Support custom homepage and about page, you can choose to use HTML & Markdown code to customize, or use a separate webpage to embed through iframe.
10. Support accessing management API through system access token.
11. Support user management, support **multiple user login and registration methods**:
     + Email login registration and password reset through email.
     + [GitHub Open License](https://github.com/settings/applications/new).
     + WeChat official account authorization (requires additional deployment of [WeChat Server](https://github.com/songquanpeng/wechat-server)).
12. In the future, after other large models open their APIs, they will be supported as soon as possible and encapsulated into the same API access method.

## deployment
### Deploy based on Docker
Execution: `docker run -d --restart always -p 3000:3000 -v /home/ubuntu/data/one-api:/data justsong/one-api`

The first `3000` in `-p 3000:3000` is the port of the host machine, which can be modified as needed.

The data will be saved in the `/home/ubuntu/data/one-api` directory of the host machine, please make sure the directory exists and has write permission, or change to a suitable directory.

Reference configuration of Nginx:
```
server {
    server_name openai.justsong.cn; # Please modify your domain name according to the actual situation
   
    location / {
           client_max_body_size 64m;
           proxy_http_version 1.1;
           proxy_pass http://localhost:3000; # Please modify your port according to the actual situation
           proxy_set_header Host $host;
           proxy_set_header X-Forwarded-For $remote_addr;
           proxy_cache_bypass $http_upgrade;
           proxy_set_header Accept-Encoding gzip;
    }
}
```

Then configure HTTPS with Let's Encrypt's certbot:
```bash
# Ubuntu install certbot:
sudo snap install --classic certbot
sudo ln -s /snap/bin/certbot /usr/bin/certbot
# Generate certificate & modify Nginx configuration
sudo certbot --nginx
# Follow instructions
# restart nginx
sudo service nginx restart
```

### Manual deployment
1. Download the executable from [GitHub Releases](https://github.com/songquanpeng/one-api/releases/latest) or compile from source:
    ```shell
    git clone https://github.com/songquanpeng/one-api.git
   
    # Build the front end
    cd one-api/web
    npm install
    npm run build

    # Build the backend
    cd..
    go mod download
    go build -ldflags "-s -w" -o one-api
    ````
2. Run:
    ```shell
    chmod u+x one-api
    ./one-api --port 3000 --log-dir ./logs
    ```
3. Visit [http://localhost:3000/](http://localhost:3000/) and log in. The initial account username is `root`, and the password is `123456`.

A more detailed deployment tutorial [see here](https://iamazing.cn/page/how-to-deploy-a-website).

### Multi-machine deployment
1. All servers `SESSION_SECRET` set the same value.
2. `SQL_DSN` must be set, use the MySQL database instead of SQLite, please configure the synchronization of the main and standby databases by yourself.
3. All slave servers must set `SYNC_FREQUENCY` to periodically synchronize configuration from the database.
4. The slave server can optionally set `FRONTEND_BASE_URL` to redirect page requests to the master server.

For details on how to use environment variables, see [here](#environment variables).

## configuration
The system itself works out of the box.

You can configure it by setting environment variables or command line parameters.

After the system starts, use `root` user to log in to the system and do further configuration.

## How to use
Add your API Key in the `Channels` page, and then add an access token in the `Tokens` page.

Then you can use your token to access the One API in the same way as [OpenAI API](https://platform.openai.com/docs/api-reference/introduction).

You can specify which channel to use to process this request by adding the channel ID after the token, for example: `Authorization: Bearer ONE_API_KEY-CHANNEL_ID`.
Note that a token created by an admin user is required to specify a channel ID.

If not added, multiple channels will be used in a load balancing manner.

### Environment variables
1. `REDIS_CONN_STRING`: After setting, Redis will be used as the storage for request frequency limit instead ofUse memory storage.
    + Example: `REDIS_CONN_STRING=redis://default:redispw@localhost:49153`
2. `SESSION_SECRET`: After setting, a fixed session key will be used, so that the logged-in user's cookie will still be valid after the system restarts.
    + Example: `SESSION_SECRET=random_string`
3. `SQL_DSN`: After setting, the specified database will be used instead of SQLite.
    + Example: `SQL_DSN=root:123456@tcp(localhost:3306)/one-api`
4. `FRONTEND_BASE_URL`: After setting, the specified front-end address will be used instead of the back-end address.
    + Example: `FRONTEND_BASE_URL=https://openai.justsong.cn`
5. `SYNC_FREQUENCY`: After setting, the configuration will be periodically synchronized with the database, in seconds, if not set, no synchronization will be performed.
    + Example: `SYNC_FREQUENCY=60`

### Command Line Arguments
1. `--port <port_number>`: Specify the port number that the server listens to, the default is `3000`.
    + Example: `--port 3000`
2. `--log-dir <log_dir>`: Specify the log folder, if not set, the log will not be saved.
    + Example: `--log-dir ./logs`
3. `--version`: Print system version number and exit.
4. `--help`: View the help and parameter description of the command.

## Demo
### Online Demo
Note that this demo site does not provide external services:
https://openai.justsong.cn

### Screenshot display
![channel](https://user-images.githubusercontent.com/39998050/233837954-ae6683aa-5c4f-429f-a949-6645a83c9490.png)
![token](https://user-images.githubusercontent.com/39998050/233837971-dab488b7-6d96-43af-b640-a168e8d1c9bf.png)

## common problem
1. Why is it prompted that the account limit is insufficient?
    + Please check whether your token limit is sufficient, this is separate from the account limit.
    + The token quota is only for the user to set the maximum usage, and the user can set it freely.
2. A blank page appears after the pagoda is deployed?
    + For automatic configuration issues, see [#97](https://github.com/songquanpeng/one-api/issues/97) for details.
