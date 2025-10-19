# Easy Sync

局域网手机-电脑互传文件与消息工具，支持断点续传、传输优化、PWA、原生移动端。

## 功能特性

- 🌐 **纯局域网传输**: 无需互联网连接，安全可靠
- 📱 **跨平台支持**: 手机浏览器、桌面浏览器均可使用
- 🔄 **断点续传**: 基于TUS协议的大文件可靠传输
- 📨 **实时消息**: WebSocket实时聊天和通知
- 🔍 **自动发现**: mDNS/Bonjour设备自动发现
- 🔐 **安全配对**: 一次性令牌安全配对机制
- 📁 **文件管理**: 完整的文件上传下载管理
- 🚀 **PWA支持**: 可安装为应用，支持离线使用

## 快速开始

### 编译和运行

```bash
# 克隆项目
git clone https://github.com/easy-sync/easy-sync.git
cd easy-sync

# 安装依赖
go mod download

# 编译
go build -o easy-sync ./cmd/server

# 运行
./easy-sync
```

### Docker运行

```bash
# 构建镜像
docker build -t easy-sync .

# 运行容器
docker run -p 3280:3280 -v $(pwd)/uploads:/app/uploads easy-sync
```

### 环境变量配置

```bash
# 服务器配置
EASY_SYNC_HOST=0.0.0.0
EASY_SYNC_PORT=3280
EASY_SYNC_HTTPS=false

# 存储配置
EASY_SYNC_UPLOAD_DIR=./uploads
EASY_SYNC_DATA_DIR=./data
EASY_SYNC_MAX_SIZE=10737418240  # 10GB

# 设备配置
EASY_SYNC_DEVICE_NAME=My-Computer

# 日志配置
EASY_SYNC_LOG_LEVEL=info
EASY_SYNC_LOG_FORMAT=json
```

## 使用方法

1. **启动服务器**: 运行Easy Sync服务器
2. **访问Web界面**: 打开浏览器访问 `http://localhost:3280`
3. **设备配对**: 扫描二维码或输入配对令牌
4. **开始传输**: 上传或下载文件，发送消息

## 项目结构

```
easy-sync/
├── cmd/server/          # 主程序入口
├── pkg/                 # 核心包
│   ├── config/         # 配置管理
│   ├── server/         # HTTP服务器
│   ├── websocket/      # WebSocket处理
│   ├── upload/         # 文件上传(TUS)
│   ├── download/       # 文件下载
│   ├── discovery/      # mDNS服务发现
│   └── security/       # 安全认证
├── web/                # 前端资源
│   └── public/         # 静态文件
└── docs/               # 文档
```

## 技术栈

### 后端
- **语言**: Go (Golang)
- **Web框架**: Gin
- **WebSocket**: gorilla/websocket
- **文件上传**: TUS协议 (tusd)
- **服务发现**: mDNS/Bonjour (grandcat/zeroconf)
- **认证**: JWT
- **日志**: logrus

### 前端
- **框架**: 原生JavaScript + HTML5
- **文件上传**: tus-js-client
- **WebSocket**: 浏览器原生
- **PWA**: manifest + service worker
- **样式**: 现代CSS (响应式设计)

## API文档

### 设备配对
- `GET /api/qr` - 获取二维码数据
- `POST /api/pair` - 完成设备配对
- `GET /api/devices` - 获取已配对设备列表
- `DELETE /api/devices/{id}` - 删除配对设备

### 文件操作
- `POST /tus/files` - 创建上传会话 (TUS协议)
- `PATCH /tus/files/{id}` - 分块上传 (TUS协议)
- `HEAD /tus/files/{id}` - 查询上传状态 (TUS协议)
- `GET /files/{id}` - 下载文件 (支持Range请求)
- `GET /files/{id}/sha256` - 获取文件校验和
- `GET /api/files` - 获取文件列表
- `DELETE /api/files/{id}` - 删除文件

### 消息通信
- `GET /api/messages` - 获取消息历史
- `WS /ws` - WebSocket实时通信

## 开发指南

### 开发环境设置

```bash
# 安装开发依赖
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行测试
go test ./...

# 代码检查
golangci-lint run

# 热重载开发
go install github.com/air-verse/air@latest
air
```

## 贡献指南

1. Fork项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建Pull Request

## 许可证

本项目采用MIT许可证。

## 原生移动端 (计划中)

Flutter/React Native，支持通知、前台保持连接、更好的文件权限