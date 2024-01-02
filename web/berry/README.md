# One API 前端界面

这个项目是 One API 的前端界面，它基于 [Berry Free React Admin Template](https://github.com/codedthemes/berry-free-react-admin-template) 进行开发。

## 使用的开源项目

使用了以下开源项目作为我们项目的一部分：

- [Berry Free React Admin Template](https://github.com/codedthemes/berry-free-react-admin-template)
- [minimal-ui-kit](minimal-ui-kit)

## 开发说明

当添加新的渠道时，需要修改以下地方：

1. `web/berry/src/constants/ChannelConstants.js`

在该文件中的 `CHANNEL_OPTIONS` 添加新的渠道

```js
export const CHANNEL_OPTIONS = {
  //key 为渠道ID
  1: {
    key: 1, // 渠道ID
    text: "OpenAI", // 渠道名称
    value: 1, // 渠道ID
    color: "primary", // 渠道列表显示的颜色
  },
};
```

2. `web/berry/src/views/Channel/type/Config.js`

在该文件中的`typeConfig`添加新的渠道配置， 如果无需配置，可以不添加

```js
const typeConfig = {
  // key 为渠道ID
  3: {
    inputLabel: {
      // 输入框名称 配置
      // 对应的字段名称
      base_url: "AZURE_OPENAI_ENDPOINT",
      other: "默认 API 版本",
    },
    prompt: {
      // 输入框提示 配置
      // 对应的字段名称
      base_url: "请填写AZURE_OPENAI_ENDPOINT",

      // 注意：通过判断 `other` 是否有值来判断是否需要显示 `other` 输入框， 默认是没有值的
      other: "请输入默认API版本，例如：2023-06-01-preview",
    },
    modelGroup: "openai", // 模型组名称,这个值是给 填入渠道支持模型 按钮使用的。 填入渠道支持模型 按钮会根据这个值来获取模型组，如果填写默认是 openai
  },
};
```

## 许可证

本项目中使用的代码遵循 MIT 许可证。
