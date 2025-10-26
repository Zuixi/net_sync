# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**easy_sync** is a local network file and message transfer tool designed for seamless communication between mobile devices and desktop computers. The project focuses on LAN-only transfers without requiring cloud services, with support for resumable uploads, PWA capabilities, and native mobile applications.

**Current Status**: ✅ **Milestone 1 完成** - 核心功能已实现并通过测试
- ✅ Go 后端服务器（HTTP + WebSocket）
- ✅ Next.js 前端界面（Tailwind v4）
- ✅ 自动设备配对系统（JWT 认证）
- ✅ WebSocket 实时通信（聊天消息）
- ✅ 多设备同时连接支持
- ✅ 文件上传功能（Tus 协议）
- 🚧 文件下载和管理功能（开发中）
- 🚧 mDNS 服务发现（需要配置）

## Technology Stack

### Backend (Go)
- **HTTP/WebSocket**: Go net/http + gorilla/websocket or fasthttp + nhooyr/ws
- **Resumable Upload**: tus-go-server (implements tus protocol for resumable file transfers)
- **Service Discovery**: github.com/grandcat/zeroconf for mDNS/Bonjour broadcasting
- **Database**: SQLite (modernc.org/sqlite) or embedded KV (bbolt/badger)
- **TLS**: Self-signed certificate generation with SHA-256 fingerprint display
- **Performance**: sendfile, Gzip/Brotli compression, HTTP/2 support

### Frontend (Next.js)
- **Framework**: Next.js for server-side rendering and static site generation
- **File Upload**: tus-js-client for resumable uploads
- **Real-time Communication**: Browser native WebSocket API
- **PWA**: Web app manifest + service worker for offline capabilities and home screen installation

### Native Mobile (Optional Enhancement)
- **Framework**: Flutter or React Native for enhanced features
- **Features**: Push notifications, foreground connection persistence, better file system permissions

## Architecture Overview

### Core Design Pattern
The project follows a **"Desktop Agent + Mobile Browser/PWA"** architecture:

1. **Desktop Agent**: Small local HTTP server running on the computer
2. **Mobile Client**: Web browser or PWA accessing the desktop agent
3. **Service Discovery**: mDNS/Bonjour for automatic device detection
4. **Security**: One-time pairing tokens with device fingerprinting

### File Transfer Flow
- **Mobile → Desktop (Upload)**: HTTP(S) + tus protocol for resumable uploads
- **Desktop → Mobile (Download)**: WebSocket file offers + HTTP range downloads
- **Bidirectional Communication**: WebSocket for real-time messages and control

### Security Model
- **Initial Pairing**: One-time tokens via QR codes or manual entry
- **Session Management**: Short-term sessions with device fingerprinting
- **Access Control**: CORS restrictions, CSRF protection, WebSocket authentication
- **Optional Encryption**: Self-signed HTTPS with certificate fingerprint verification

## Development Commands

### Backend (Go)

**开发模式启动（推荐）**
```bash
# 使用配置文件启动服务器（默认端口 3280）
go run cmd/server/main.go -config config.yaml

# 使用默认配置启动
go run cmd/server/main.go

# 查看版本信息
go run cmd/server/main.go -version
```

**构建和部署**
```bash
# 构建当前平台的可执行文件
go build -o easy-sync ./cmd/server

# 跨平台构建
GOOS=windows GOARCH=amd64 go build -o easy-sync.exe ./cmd/server
GOOS=darwin GOARCH=arm64 go build -o easy-sync-mac ./cmd/server
GOOS=linux GOARCH=amd64 go build -o easy-sync-linux ./cmd/server

# 运行构建的可执行文件
./easy-sync -config config.yaml
```

**测试**
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./pkg/websocket
go test ./pkg/security
```

### Frontend (Next.js)

**重要**: 前端代码位于 `web/client` 目录下

```bash
# 进入前端目录
cd web/client

# 首次运行需要安装依赖
npm install

# 启动开发服务器（默认 http://localhost:3000，如被占用会自动使用其他端口）
npm run dev

# 构建生产版本
npm run build

# 运行生产版本
npm start
```

**端到端测试（使用 Playwright）**
```bash
# 确保后端服务器已启动（端口 3280）
# 确保前端开发服务器已启动

# 运行 Playwright 测试
npx playwright test

# 运行特定测试文件
npx playwright test tests/chat-message.spec.ts

# 以 UI 模式运行测试
npx playwright test --ui

# 查看测试报告
npx playwright show-report
```

### 完整开发环境启动流程

1. **启动后端服务器**（终端 1）
```bash
go run cmd/server/main.go -config config.yaml
```
预期输出：服务器运行在 `http://[::]:3280`，并显示配对令牌

2. **启动前端开发服务器**（终端 2）
```bash
cd web/client
npm run dev
```
预期输出：服务器运行在 `http://localhost:3000` 或其他可用端口

3. **访问应用**
在浏览器中打开前端地址（如 `http://localhost:3003`），应用会自动进行设备配对并建立 WebSocket 连接

## Key API Endpoints

### 系统端点
- ✅ `GET /health` - 健康检查
- ✅ `GET /api/config` - 获取服务器配置信息

### 设备配对与发现 ✅ 已实现
- ✅ `GET /api/qr` - 获取配对二维码
- ✅ `GET /api/pairing-token` - 获取配对令牌（用于自动配对）
- ✅ `POST /api/pair` - 完成设备配对（返回 JWT token）
- ✅ `GET /api/devices` - 列出已配对设备
- ✅ `DELETE /api/devices/:id` - 移除设备配对

### WebSocket 通信 ✅ 已实现
- ✅ `GET /ws` - WebSocket 端点（需要 JWT token 认证）
  - 支持实时聊天消息
  - 支持连接状态广播
  - 支持多客户端消息转发

### 文件操作 ✅ 部分实现
- ✅ `POST /tus/*filepath` - 创建上传会话（tus 协议）
- ✅ `PATCH /tus/*filepath` - 分块上传（tus 协议）
- ✅ `HEAD /tus/*filepath` - 查询上传偏移量（tus 协议）
- ✅ `GET /tus/*filepath` - 获取上传信息
- ✅ `OPTIONS /tus/*filepath` - CORS 预检请求
- ✅ `GET /files/:id` - 下载文件（支持 Range 请求）
- ✅ `GET /files/:id/sha256` - 获取文件校验和
- ✅ `GET /api/files` - 列出文件及元数据
- ✅ `DELETE /api/files/:id` - 删除文件

### 消息历史 ✅ 已实现
- ✅ `GET /api/messages?before&limit` - 分页获取历史消息

## WebSocket Message Types

### Client → Server
```json
{
  "type": "hello",
  "device": "iPhone-15-Pro",
  "capabilities": ["upload", "download", "chat"]
}

{
  "type": "chat",
  "id": "c-1731",
  "text": "Message content",
  "timestamp": 1731672000
}

{
  "type": "file_offer_ack",
  "offer_id": "f-991",
  "accepted": true
}
```

### Server → Client
```json
{
  "type": "chat",
  "id": "c-1732",
  "from": "My-PC",
  "text": "Reply message",
  "timestamp": 1731672003
}

{
  "type": "file_offer",
  "offer_id": "f-991",
  "from": "My-PC",
  "name": "video.mp4",
  "size": 2147483648,
  "mime": "video/mp4",
  "sha256": "...",
  "url": "https://host:port/files/f-991?sig=...&exp=..."
}
```

## Performance Considerations

### Large File Support
- **Zero-copy transfers**: Use `sendfile` system calls
- **Concurrent downloads**: Range request support for parallel downloads
- **Memory efficiency**: Stream processing without loading entire files into memory

### Network Reliability
- **Resumable uploads**: Tus protocol handles connection interruptions
- **Automatic reconnection**: WebSocket reconnection with exponential backoff
- **Bandwidth optimization**: Compression for text-based content

### Cross-platform Compatibility
- **Service discovery fallbacks**: Manual IP entry when mDNS is unavailable
- **Enterprise network support**: Custom ports and HTTP-only modes
- **Device-specific optimizations**: iOS file picker limitations, Android storage access

## Implementation Roadmap

### Milestone 1: Basic Functionality (1-2 days)
- HTTP server with Web UI
- Simple file upload/download
- Basic WebSocket chat

### Milestone 2: Reliable Transfer (2-4 days)
- Tus protocol implementation for resumable uploads
- Range download support
- File metadata and checksum validation

### Milestone 3: Discovery & Security (3-5 days)
- mDNS service discovery
- One-time pairing system
- Self-signed HTTPS support

### Milestone 4: Enhanced Features
- PWA capabilities
- Desktop tray integration
- Performance optimizations

## Testing Strategy

### Platform Matrix
- **Mobile**: iOS Safari, iOS Chrome, Android Chrome
- **Desktop**: Windows (Chrome, Edge, Firefox), macOS (Safari, Chrome, Firefox), Linux (Chrome, Firefox)

### Test Scenarios
- File types and sizes (small files to >10GB)
- Network conditions (weak WiFi, AP switching, connection drops)
- Security workflows (pairing, token expiration, certificate verification)
- Service discovery (mDNS enabled/disabled scenarios)

## Test Results

### 最新测试 (2025-10-26)

**测试环境**
- 后端: Go 服务器运行在 Windows 10，端口 3280
- 前端: Next.js 15.5.6 (Turbopack)，端口 3003
- 测试工具: Playwright MCP

**测试结果** ✅ 全部通过

| 测试项目 | 状态 | 详情 |
|---------|------|------|
| 服务器启动 | ✅ 通过 | 所有 API 端点正常注册 |
| 前端加载 | ✅ 通过 | 页面正常渲染，自动配对成功 |
| 设备配对 | ✅ 通过 | JWT token 生成和验证正常 |
| WebSocket 连接 | ✅ 通过 | 连接稳定，状态同步正常 |
| 聊天消息 | ✅ 通过 | 消息发送和广播功能正常 |
| 多客户端 | ✅ 通过 | 支持多设备同时连接 |

**已知问题**
1. mDNS 服务发现需要配置设备名称（`mdns.device_name`）
2. Next.js 开发模式的热重载会导致 WebSocket 短暂断开

**测试截图**: `.playwright-mcp/chat-test-success.png`

## Important Notes

- **LAN-only operation**: No cloud dependencies or internet connectivity required
- **Security-first approach**: All transfers require device pairing and authentication
- **Cross-platform focus**: Prioritize browser compatibility over native features initially
- **Progressive enhancement**: Start with basic HTTP transfers, add advanced features incrementally
- network request,should set http proxy first:
`$env:HTTP_PROXY="http://127.0.0.1:13824"`
`$env:HTTPS_PROXY="http://127.0.0.1:13824"`
- 总是用中文回答
- Blank lines should not contain any spaces.