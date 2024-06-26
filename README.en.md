<p align="right">
    <a href="./README.md">中文</a> | <strong>English</strong> | <a href="./README.ja.md">日本語</a>
</p>

<p align="center">
  <a href="https://github.com/songquanpeng/one-api"><img src="https://raw.githubusercontent.com/songquanpeng/one-api/main/web/default/public/logo.png" width="150" height="150" alt="one-api logo"></a>
</p>

<div align="center">

# One API

_✨ Access all LLM through the standard OpenAI API format, easy to deploy & use ✨_

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
  <a href="#deployment">Deployment Tutorial</a>
  ·
  <a href="#usage">Usage</a>
  ·
  <a href="https://github.com/songquanpeng/one-api/issues">Feedback</a>
  ·
  <a href="#screenshots">Screenshots</a>
  ·
  <a href="https://openai.justsong.cn/">Live Demo</a>
  ·
  <a href="#faq">FAQ</a>
  ·
  <a href="#related-projects">Related Projects</a>
  ·
  <a href="https://iamazing.cn/page/reward">Donate</a>
</p>

> **Warning**: This README is translated by ChatGPT. Please feel free to submit a PR if you find any translation errors.

> **Warning**: The Docker image for English version is `justsong/one-api-en`.

> **Note**: The latest image pulled from Docker may be an `alpha` release. Specify the version manually if you require stability.

## Features
1. Support for multiple large models:
   + [x] [OpenAI ChatGPT Series Models](https://platform.openai.com/docs/guides/gpt/chat-completions-api) (Supports [Azure OpenAI API](https://learn.microsoft.com/en-us/azure/ai-services/openai/reference))
   + [x] [Anthropic Claude Series Models](https://anthropic.com)
   + [x] [Google PaLM2 and Gemini Series Models](https://developers.generativeai.google)
   + [x] [Baidu Wenxin Yiyuan Series Models](https://cloud.baidu.com/doc/WENXINWORKSHOP/index.html)
   + [x] [Alibaba Tongyi Qianwen Series Models](https://help.aliyun.com/document_detail/2400395.html)
   + [x] [Zhipu ChatGLM Series Models](https://bigmodel.cn)
2. Supports access to multiple channels through **load balancing**.
3. Supports **stream mode** that enables typewriter-like effect through stream transmission.
4. Supports **multi-machine deployment**. [See here](#multi-machine-deployment) for more details.
5. Supports **token management** that allows setting token expiration time and usage count.
6. Supports **voucher management** that enables batch generation and export of vouchers. Vouchers can be used for account balance replenishment.
7. Supports **channel management** that allows bulk creation of channels.
8. Supports **user grouping** and **channel grouping** for setting different rates for different groups.
9. Supports channel **model list configuration**.
10. Supports **quota details checking**.
11. Supports **user invite rewards**.
12. Allows display of balance in USD.
13. Supports announcement publishing, recharge link setting, and initial balance setting for new users.
14. Offers rich **customization** options:
    1. Supports customization of system name, logo, and footer.
    2. Supports customization of homepage and about page using HTML & Markdown code, or embedding a standalone webpage through iframe.
15. Supports management API access through system access tokens.
16. Supports Cloudflare Turnstile user verification.
17. Supports user management and multiple user login/registration methods:
    + Email login/registration and password reset via email.
    + [GitHub OAuth](https://github.com/settings/applications/new).
    + WeChat Official Account authorization (requires additional deployment of [WeChat Server](https://github.com/songquanpeng/wechat-server)).
18. Immediate support and encapsulation of other major model APIs as they become available.

## Deployment
### Docker Deployment
Deployment command: `docker run --name one-api -d --restart always -p 3000:3000 -e TZ=Asia/Shanghai -v /home/ubuntu/data/one-api:/data justsong/one-api-en`

Update command: `docker run --rm -v /var/run/docker.sock:/var/run/docker.sock containrrr/watchtower -cR`

The first `3000` in `-p 3000:3000` is the port of the host, which can be modified as needed.

Data will be saved in the `/home/ubuntu/data/one-api` directory on the host. Ensure that the directory exists and has write permissions, or change it to a suitable directory.

Nginx reference configuration:
```
server{
   server_name openai.justsong.cn;  # Modify your domain name accordingly

   location / {
          client_max_body_size  64m;
          proxy_http_version 1.1;
          proxy_pass http://localhost:3000;  # Modify your port accordingly
          proxy_set_header Host $host;
          proxy_set_header X-Forwarded-For $remote_addr;
          proxy_cache_bypass $http_upgrade;
          proxy_set_header Accept-Encoding gzip;
   }
}
```

Next, configure HTTPS with Let's Encrypt certbot:
```bash
# Install certbot on Ubuntu:
sudo snap install --classic certbot
sudo ln -s /snap/bin/certbot /usr/bin/certbot
# Generate certificates & modify Nginx configuration
sudo certbot --nginx
# Follow the prompts
# Restart Nginx
sudo service nginx restart
```

The initial account username is `root` and password is `123456`.

### Manual Deployment
1. Download the executable file from [GitHub Releases](https://github.com/songquanpeng/one-api/releases/latest) or compile from source:
   ```shell
   git clone https://github.com/songquanpeng/one-api.git

   # Build the frontend
   cd one-api/web/default
   npm install
   npm run build

   # Build the backend
   cd ../..
   go mod download
   go build -ldflags "-s -w" -o one-api
   ```
2. Run:
   ```shell
   chmod u+x one-api
   ./one-api --port 3000 --log-dir ./logs
   ```
3. Access [http://localhost:3000/](http://localhost:3000/) and log in. The initial account username is `root` and password is `123456`.

For more detailed deployment tutorials, please refer to [this page](https://iamazing.cn/page/how-to-deploy-a-website).

### Multi-machine Deployment
1. Set the same `SESSION_SECRET` for all servers.
2. Set `SQL_DSN` and use MySQL instead of SQLite. All servers should connect to the same database.
3. Set the `NODE_TYPE` for all non-master nodes to `slave`.
4. Set `SYNC_FREQUENCY` for servers to periodically sync configurations from the database.
5. Non-master nodes can optionally set `FRONTEND_BASE_URL` to redirect page requests to the master server.
6. Install Redis separately on non-master nodes, and configure `REDIS_CONN_STRING` so that the database can be accessed with zero latency when the cache has not expired.
7. If the main server also has high latency accessing the database, Redis must be enabled and `SYNC_FREQUENCY` must be set to periodically sync configurations from the database.

Please refer to the [environment variables](#environment-variables) section for details on using environment variables.

### Deployment on Control Panels (e.g., Baota)
Refer to [#175](https://github.com/songquanpeng/one-api/issues/175) for detailed instructions.

If you encounter a blank page after deployment, refer to [#97](https://github.com/songquanpeng/one-api/issues/97) for possible solutions.

### Deployment on Third-Party Platforms
<details>
<summary><strong>Deploy on Sealos</strong></summary>
<div>

> Sealos supports high concurrency, dynamic scaling, and stable operations for millions of users.

> Click the button below to deploy with one click.👇

[![](https://raw.githubusercontent.com/labring-actions/templates/main/Deploy-on-Sealos.svg)](https://cloud.sealos.io/?openapp=system-fastdeploy?templateName=one-api)


</div>
</details>

<details>
<summary><strong>Deployment on Zeabur</strong></summary>
<div>

> Zeabur's servers are located overseas, automatically solving network issues, and the free quota is sufficient for personal usage.

[![Deploy on Zeabur](https://zeabur.com/button.svg)](https://zeabur.com/templates/7Q0KO3)

1. First, fork the code.
2. Go to [Zeabur](https://zeabur.com?referralCode=songquanpeng), log in, and enter the console.
3. Create a new project. In Service -> Add Service, select Marketplace, and choose MySQL. Note down the connection parameters (username, password, address, and port).
4. Copy the connection parameters and run ```create database `one-api` ``` to create the database.
5. Then, in Service -> Add Service, select Git (authorization is required for the first use) and choose your forked repository.
6. Automatic deployment will start, but please cancel it for now. Go to the Variable tab, add a `PORT` with a value of `3000`, and then add a `SQL_DSN` with a value of `<username>:<password>@tcp(<addr>:<port>)/one-api`. Save the changes. Please note that if `SQL_DSN` is not set, data will not be persisted, and the data will be lost after redeployment.
7. Select Redeploy.
8. In the Domains tab, select a suitable domain name prefix, such as "my-one-api". The final domain name will be "my-one-api.zeabur.app". You can also CNAME your own domain name.
9. Wait for the deployment to complete, and click on the generated domain name to access One API.

</div>
</details>

## Configuration
The system is ready to use out of the box.

You can configure it by setting environment variables or command line parameters.

After the system starts, log in as the `root` user to further configure the system.

## Usage
Add your API Key on the `Channels` page, and then add an access token on the `Tokens` page.

You can then use your access token to access One API. The usage is consistent with the [OpenAI API](https://platform.openai.com/docs/api-reference/introduction).

In places where the OpenAI API is used, remember to set the API Base to your One API deployment address, for example: `https://openai.justsong.cn`. The API Key should be the token generated in One API.

Note that the specific API Base format depends on the client you are using.

```mermaid
graph LR
    A(User)
    A --->|Request| B(One API)
    B -->|Relay Request| C(OpenAI)
    B -->|Relay Request| D(Azure)
    B -->|Relay Request| E(Other downstream channels)
```

To specify which channel to use for the current request, you can add the channel ID after the token, for example: `Authorization: Bearer ONE_API_KEY-CHANNEL_ID`.
Note that the token needs to be created by an administrator to specify the channel ID.

If the channel ID is not provided, load balancing will be used to distribute the requests to multiple channels.

### Environment Variables
1. `REDIS_CONN_STRING`: When set, Redis will be used as the storage for request rate limiting instead of memory.
    + Example: `REDIS_CONN_STRING=redis://default:redispw@localhost:49153`
2. `SESSION_SECRET`: When set, a fixed session key will be used to ensure that cookies of logged-in users are still valid after the system restarts.
    + Example: `SESSION_SECRET=random_string`
3. `SQL_DSN`: When set, the specified database will be used instead of SQLite. Please use MySQL version 8.0.
    + Example: `SQL_DSN=root:123456@tcp(localhost:3306)/oneapi`
4. `LOG_SQL_DSN`: When set, a separate database will be used for the `logs` table; please use MySQL or PostgreSQL.
    + Example: `LOG_SQL_DSN=root:123456@tcp(localhost:3306)/oneapi-logs`
5. `FRONTEND_BASE_URL`: When set, the specified frontend address will be used instead of the backend address.
    + Example: `FRONTEND_BASE_URL=https://openai.justsong.cn`
6. `SYNC_FREQUENCY`: When set, the system will periodically sync configurations from the database, with the unit in seconds. If not set, no sync will happen.
    + Example: `SYNC_FREQUENCY=60`
7. `NODE_TYPE`: When set, specifies the node type. Valid values are `master` and `slave`. If not set, it defaults to `master`.
    + Example: `NODE_TYPE=slave`
8. `CHANNEL_UPDATE_FREQUENCY`: When set, it periodically updates the channel balances, with the unit in minutes. If not set, no update will happen.
    + Example: `CHANNEL_UPDATE_FREQUENCY=1440`
9. `CHANNEL_TEST_FREQUENCY`: When set, it periodically tests the channels, with the unit in minutes. If not set, no test will happen.
    + Example: `CHANNEL_TEST_FREQUENCY=1440`
10. `POLLING_INTERVAL`: The time interval (in seconds) between requests when updating channel balances and testing channel availability. Default is no interval.
    + Example: `POLLING_INTERVAL=5`

### Command Line Parameters
1. `--port <port_number>`: Specifies the port number on which the server listens. Defaults to `3000`.
    + Example: `--port 3000`
2. `--log-dir <log_dir>`: Specifies the log directory. If not set, the logs will not be saved.
    + Example: `--log-dir ./logs`
3. `--version`: Prints the system version number and exits.
4. `--help`: Displays the command usage help and parameter descriptions.

## Screenshots
![channel](https://user-images.githubusercontent.com/39998050/233837954-ae6683aa-5c4f-429f-a949-6645a83c9490.png)
![token](https://user-images.githubusercontent.com/39998050/233837971-dab488b7-6d96-43af-b640-a168e8d1c9bf.png)

## FAQ
1. What is quota? How is it calculated? Does One API have quota calculation issues?
    + Quota = Group multiplier * Model multiplier * (number of prompt tokens + number of completion tokens * completion multiplier)
    + The completion multiplier is fixed at 1.33 for GPT3.5 and 2 for GPT4, consistent with the official definition.
    + If it is not a stream mode, the official API will return the total number of tokens consumed. However, please note that the consumption multipliers for prompts and completions are different.
2. Why does it prompt "insufficient quota" even though my account balance is sufficient?
    + Please check if your token quota is sufficient. It is separate from the account balance.
    + The token quota is used to set the maximum usage and can be freely set by the user.
3. It says "No available channels" when trying to use a channel. What should I do?
    + Please check the user and channel group settings.
    + Also check the channel model settings.
4. Channel testing reports an error: "invalid character '<' looking for beginning of value"
    + This error occurs when the returned value is not valid JSON but an HTML page.
    + Most likely, the IP of your deployment site or the node of the proxy has been blocked by CloudFlare.
5. ChatGPT Next Web reports an error: "Failed to fetch"
    + Do not set `BASE_URL` during deployment.
    + Double-check that your interface address and API Key are correct.

## Related Projects
* [FastGPT](https://github.com/labring/FastGPT): Knowledge question answering system based on the LLM
* [VChart](https://github.com/VisActor/VChart):  More than just a cross-platform charting library, but also an expressive data storyteller.
* [VMind](https://github.com/VisActor/VMind):  Not just automatic, but also fantastic. Open-source solution for intelligent visualization.

## Note
This project is an open-source project. Please use it in compliance with OpenAI's [Terms of Use](https://openai.com/policies/terms-of-use) and **applicable laws and regulations**. It must not be used for illegal purposes.

This project is released under the MIT license. Based on this, attribution and a link to this project must be included at the bottom of the page.

The same applies to derivative projects based on this project.

If you do not wish to include attribution, prior authorization must be obtained.

According to the MIT license, users should bear the risk and responsibility of using this project, and the developer of this open-source project is not responsible for this.
