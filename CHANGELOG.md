## [2024-01-01]

### 优化
- **Dockerfile**: 优化 Docker 构建，配置国内镜像加速
  - Node.js 构建阶段：配置 npm 使用 npmmirror 镜像源加速依赖下载
  - Go 构建阶段：替换 Alpine apk 为阿里云镜像加速系统包安装，配置 Go 模块代理为 goproxy.cn
  - 运行时镜像：固定 Alpine 版本为 3.18，添加时区数据配置（Asia/Shanghai）
  - 升级基础镜像：node:16 -> node:18-alpine，golang:alpine -> golang:1.22-alpine
