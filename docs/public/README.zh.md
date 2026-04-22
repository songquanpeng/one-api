# DerEineRing（ZH）

## 项目简介
DerEineRing 是基于 OneAPI 的公开定制版本，重点是高质量多语言前端体验。

核心方向：
- 多语言 UI：中文、英文、德文、塞尔维亚文
- 清晰的公开文档与翻译工作流
- 前端界面与交互体验优化

## 适用人群
- 本地部署 OneAPI 网关的团队
- 需要多语言管理界面的用户
- 参与 i18n 与前端优化的贡献者

## 已支持语言
- `zh`
- `en`
- `de`
- `sr`

语言文件：
- `web/default/src/locales/zh/translation.json`
- `web/default/src/locales/en/translation.json`
- `web/default/src/locales/de/translation.json`
- `web/default/src/locales/sr/translation.json`

## 快速安装（Docker）
```bash
git clone https://github.com/olivilo/one-api.git
cd one-api
docker compose up -d --build
```

访问地址：
- `http://localhost:3000`

## 首次登录
在全新数据库下，OneAPI 通常会创建：
- 用户名：`root`
- 密码：`123456`

首次登录后请立即修改密码。

## 使用方式
1. 登录 Web 界面。
2. 在 **Channels** 中添加提供商/渠道。
3. 在 **Tokens** 中创建访问令牌。
4. 将 OneAPI 作为 OpenAI 兼容接口使用：
   - Base URL：`http://localhost:3000/v1`
   - API Key：Web 界面中创建的 token

调用示例：
```bash
curl http://localhost:3000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ONEAPI_TOKEN" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role":"user","content":"你好，来自 DerEineRing"}]
  }'
```

## 界面语言与主题
- 在页面顶部语言菜单切换 `zh/en/de/sr`。
- 在页面顶部主题菜单切换可用主题。

## 贡献说明
- 文本修改请在对应 locale JSON 文件中进行。
- 四种语言尽量保持术语一致。
- 通过 Pull Request 提交翻译与 UI/UX 改进。
