<p align="center">
   <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://github.com/MartialBE/one-api/assets/42402987/c4125d1a-5577-446d-ba15-2a71c52140c1">
   <img height="90" src="https://raw.githubusercontent.com/MartialBE/one-api/main/web/src/assets/images/logo.svg">
   </picture>
</p>

<div align="center">

# One API

_本项目是基于[one-api](https://github.com/songquanpeng/one-api)二次开发而来的，主要将原项目中的模块代码分离，模块化，并修改了前端界面。本项目同样遵循 MIT 协议。_

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
  <a href="https://goreportcard.com/report/github.com/MartialBE/one-api">
    <img src="https://goreportcard.com/badge/github.com/MartialBE/one-api" alt="GoReportCard">
  </a>
</p>

**请不要和原版混用，因为新增功能，数据库与原版不兼容**

[演示网站](https://one-api-martialbe.vercel.app/)

</div>

## 功能变化

- 全新的 UI 界面
- 新增用户仪表盘
- 新增管理员分析数据统计界面
- 重构了中转`供应商`模块
- 支持使用`Azure Speech`模拟`TTS`功能
- 渠道可配置单独的 http/socks5 代理
- 支持动态返回用户模型列表
- 支持自定义测速模型
- 日志增加请求耗时
- 支持和优化非 OpenAI 模型的函数调用（支持的模型可以在 lobe-chat 直接使用）
- 支持完成倍率自定义
- 支持完整的分页和排序
- 支持`Telegram bot`

## 文档

请查看[文档](https://github.com/MartialBE/one-api/wiki)

## 其他

<a href="https://next.ossinsight.io/widgets/official/analyze-repo-stars-history?repo_id=689214770" target="_blank" style="display: block" align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://next.ossinsight.io/widgets/official/analyze-repo-stars-history/thumbnail.png?repo_id=689214770&image_size=auto&color_scheme=dark" width="721" height="auto">
    <img alt="Star History of MartialBE/one-api" src="https://next.ossinsight.io/widgets/official/analyze-repo-stars-history/thumbnail.png?repo_id=689214770&image_size=auto&color_scheme=light" width="721" height="auto">
  </picture>
</a>
