<p align="right">
   <strong>English</strong> | <a href="./README.md">中文</a>
</p>

<p align="center">
   <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://github.com/MartialBE/one-api/assets/42402987/c4125d1a-5577-446d-ba15-2a71c52140c1">
   <img height="90" src="https://raw.githubusercontent.com/MartialBE/one-api/main/web/src/assets/images/logo.svg">
   </picture>
</p>

<div align="center">

# One API

_This project is based on [one-api](https://github.com/songquanpeng/one-api) and has been developed for the second time. The main purpose is to separate the module code in the original project, modularize it, and modify the front-end interface. This project also follows the MIT protocol._

<p align="center">
  <a href="https://raw.githubusercontent.com/MartialBE/one-api/main/LICENSE">
    <img src="https://img.shields.io/github/license/MartialBE/one-api?color=brightgreen" alt="license">
  </a>
  <a href="https://github.com/MartialBE/one-api/releases/latest">
    <img src="https://img.shields.io/github/v/release/MartialBE/one-api?color=brightgreen&include_prereleases" alt="release">
  </a>
  <a href="https://github.com/users/MartialBE/packages/container/package/one-api">
    <img src="https://img.shields.io/badge/docker-ghcr.io-blue" alt="docker">
  </a>
  <a href="https://hub.docker.com/r/martialbe/one-api">
    <img src="https://img.shields.io/badge/docker-dockerHub-blue" alt="docker">
  </a>
  <a href="https://goreportcard.com/report/github.com/MartialBE/one-api">
    <img src="https://goreportcard.com/badge/github.com/MartialBE/one-api" alt="GoReportCard">
  </a>
</p>

**Please do not mix with the original version, because the new functions, the database is not compatible with the original version**

**For the sake of simplicity, after this project, except for updating the model list built into the program when adding a new supplier, the model list built into the program will not be updated under normal circumstances.**

If you find that a new model is missing, please update the newly added model in `Backend-Model Price-Update Price`

[Demo Site](https://one-api-martialbe.vercel.app/)

</div>

## Functional Changes

- Brand new UI interface
- Added user dashboard
- Added administrator data analysis and statistics interface
- Refactored the intermediary `supplier` module
- Support for using `Azure Speech` to simulate `TTS` function
- Channels can be configured with separate http/socks5 proxies
- Support for dynamically returning user model lists
- Support for custom speed testing models
- Logs now include request duration
- Support and optimize function calls for non-OpenAI models (supported models can be used directly in Lobe-Chat)
- Support for custom completion rates
- Support for full pagination and sorting
- Support for `Telegram bot`
- Support for models charged per use
- Support for model wildcards
- Support for starting the program using a configuration file

## Documentation

Please refer to the [documentation](https://github.com/MartialBE/one-api/wiki).

## Current Supported Providers

| Provider                                                               | Chat                     | Embeddings | Audio  | Images      | Other                                                             |
| --------------------------------------------------------------------- | ------------------------ | ---------- | ------ | ----------- | ---------------------------------------------------------------- |
| [OpenAI](https://platform.openai.com/docs/api-reference/introduction) | ✅                       | ✅         | ✅     | ✅          | -                                                                |
| [Azure OpenAI](https://oai.azure.com/)                                | ✅                       | ✅         | ✅     | ✅          | -                                                                |
| [Azure Speech](https://portal.azure.com/)                             | -                        | -          | ⚠️ tts | -           | -                                                                |
| [Anthropic](https://www.anthropic.com/)                               | ✅                       | -          | -      | -           | -                                                                |
| [Gemini](https://aistudio.google.com/)                                | ✅                       | -          | -      | -           | -                                                                |
| [百度文心](https://console.bce.baidu.com/qianfan/overview)            | ✅                       | ✅         | -      | -           | -                                                                |
| [通义千问](https://dashscope.console.aliyun.com/overview)             | ✅                       | ✅         | -      | -           | -                                                                |
| [讯飞星火](https://console.xfyun.cn/)                                 | ✅                       | -          | -      | -           | -                                                                |
| [智谱](https://open.bigmodel.cn/overview)                             | ✅                       | ✅         | -      | ⚠️ image | -                                                                |
| [腾讯混元](https://cloud.tencent.com/product/hunyuan)                 | ✅                       | -          | -      | -           | -                                                                |
| [百川](https://platform.baichuan-ai.com/console/apikey)               | ✅                       | ✅         | -      | -           | -                                                                |
| [MiniMax](https://www.minimaxi.com/user-center/basic-information)     | ✅                       | ✅         | -      | -           | -                                                                |
| [Deepseek](https://platform.deepseek.com/usage)                       | ✅                       | -          | -      | -           | -                                                                |
| [Moonshot](https://moonshot.ai/)                                      | ✅                       | -          | -      | -           | -                                                                |
| [Mistral](https://mistral.ai/)                                        | ✅                       | ✅         | -      | -           | -                                                                |
| [Groq](https://console.groq.com/keys)                                 | ✅                       | -          | -      | -           | -                                                                |
| [Amazon Bedrock](https://console.aws.amazon.com/bedrock/home)         | ⚠️ Only support Anthropic models | -          | -      | -           | -                                                                |
| [零一万物](https://platform.lingyiwanwu.com/details)                  | ✅                       | -          | -      | -           | -                                                                |
| [Cloudflare AI](https://ai.cloudflare.com/)                           | ✅                       | -          | ⚠️ stt | ⚠️ image | -                                                                |
| [Midjourney](https://www.midjourney.com/)                             | -                        | -          | -      | -           | [midjourney-proxy](https://github.com/novicezk/midjourney-proxy) |
| [Cohere](https://cohere.com/)                                         | ✅                       | -          | -      | -           | -                                                                |
| [Stability AI](https://platform.stability.ai/account/credits)         | -                        | -          | -      | ⚠️ image | -                                                                |
| [Coze](https://www.coze.com/open/docs/chat?_lang=zh)                  | ✅                       | -          | -      | -           | -                                                                |

## Acknowledgements

- This program utilizes the following open-source projects:
  - [one-api](https://github.com/songquanpeng/one-api) serves as the foundation of this project.
  - [Berry Free React Admin Template](https://github.com/codedthemes/berry-free-react-admin-template) provides the frontend interface for this project.
  - [minimal-ui-kit](https://github.com/minimal-ui-kit/material-kit-react), some styles from this project were used.
  - [new api](https://github.com/Calcium-Ion/new-api), the code for the Midjourney module is sourced from here.

Special thanks to the authors and contributors of the above projects.

## Others

<a href="https://next.ossinsight.io/widgets/official/analyze-repo-stars-history?repo_id=689214770" target="_blank" style="display: block" align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://next.ossinsight.io/widgets/official/analyze-repo-stars-history/thumbnail.png?repo_id=689214770&image_size=auto&color_scheme=dark" width="721" height="auto">
    <img alt="Star History of MartialBE/one-api" src="https://next.ossinsight.io/widgets/official/analyze-repo-stars-history/thumbnail.png?repo_id=689214770&image_size=auto&color_scheme=light" width="721" height="auto">
  </picture>
</a>
