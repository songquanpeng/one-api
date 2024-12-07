# 代码仓库结构

## 文档目录说明：
```
.
├── README.md                   - README
├── docs                        - 服务文档相关文件
│   └── index.md
├── resources                   - 服务资源文件
│   ├── icons
│   │   └── service_logo.png    - 服务logo
│   └── artifact_resources      - 部署物相关资源文件
├── ros_templates               - 服务ROS模板目录，支持多模板
│   └── template.yaml           - ROS模板，ROS模板引擎根据该模板会自动创建出所有的资源
├── config.yaml                 - 服务配置文件，服务构建过程中会使用计算巢命令行工具computenest-cli，computenest-cli会基于该配置文件构建服务
├── preset_parameters.yaml      - （该文件只有托管版有）服务商预设参数，如VpcId，VSwitchId等，该ros模板内容会渲染为表单方便服务商填写
```

## 其他
关于ROS模板，请参见 [资源编排](https://help.aliyun.com/zh/ros)。
关于computenest-cli请参见 [computenest-cli](https://pypi.org/project/computenest-cli/)。