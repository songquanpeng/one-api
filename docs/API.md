# 使用 API 操控 & 扩展 One API
> 欢迎提交 PR 在此放上你的拓展项目。

例如，虽然 One API 本身没有直接支持支付，但是你可以通过系统扩展的 API 来实现支付功能。

又或者你想自定义渠道管理策略，也可以通过 API 来实现渠道的禁用与启用。

## 鉴权
One API 支持两种鉴权方式：Cookie 和 Token，对于 Token，参照下图获取：

![image](https://github.com/songquanpeng/songquanpeng.github.io/assets/39998050/c15281a7-83ed-47cb-a1f6-913cb6bf4a7c)

之后，将 Token 作为请求头的 Authorization 字段的值即可，例如下面使用 Token 调用测试渠道的 API：
![image](https://github.com/songquanpeng/songquanpeng.github.io/assets/39998050/1273b7ae-cb60-4c0d-93a6-b1cbc039c4f8)

## 请求格式与响应格式
One API 使用 JSON 格式进行请求和响应。

对于响应体，一般格式如下：
```json
{
  "message": "请求信息",
  "success": true,
  "data": {}
}
```

## API 列表
> 当前 API 列表不全，请自行通过浏览器抓取前端请求

如果现有的 API 没有办法满足你的需求，欢迎提交 issue 讨论。

### 获取当前登录用户信息
**GET** `/api/user/self`

### 为给定用户充值额度
**POST** `/api/topup`
```json
{
  "user_id": 1,
  "quota": 100000,
  "remark": "充值 100000 额度"
}
```

## 其他
### 充值链接上的附加参数
One API 会在用户点击充值按钮的时候，将用户的信息和充值信息附加在链接上，例如：
`https://example.com?username=root&user_id=1&transaction_id=4b3eed80-55d5-443f-bd44-fb18c648c837`

你可以通过解析链接上的参数来获取用户信息和充值信息，然后调用 API 来为用户充值。

注意，不是所有主题都支持该功能，欢迎 PR 补齐。