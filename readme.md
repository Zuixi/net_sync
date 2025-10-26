# Easy Sync

局域网手机-电脑互传文件与消息工具，支持断点续传、传输优化、PWA、原生移动端（计划）。

## 功能特性
- 🌐 纯局域网传输：无需互联网连接，安全可靠
- 📱 跨平台支持：手机/桌面浏览器均可使用
- 🔄 断点续传：基于 TUS 协议可靠上传大文件
- 📨 实时消息：WebSocket 聊天与通知
- 🔍 自动发现：mDNS/Bonjour 设备发现
- 🔐 安全配对：一次性令牌安全配对
- 📁 文件管理：上传、下载、校验、删除
- 🚀 PWA：可安装为应用，支持离线（Next 已集成 manifest，SW 集成待完善）

---

## Web 架构设计

- 技术栈
  - 前端：Next.js 15（App Router, TypeScript）+ Tailwind v4 + Heroicons + PWA manifest
  - 上传：`tus-js-client`（断点续传，/tus/files）
  - 实时通信：浏览器原生 WebSocket（`/ws`）
  - 后端：Go + Gin，文件上传（tusd handler），下载与校验、设备与消息管理

- 模块划分（`web/client/src`）
  - `components/StatusBar.tsx`：连接状态指示
  - `components/Pairing.tsx`：输入令牌，`POST /api/pair` 完成配对，保存返回 token
  - `components/Uploader.tsx`：TUS 上传（进度与速率显示），成功后提供验证与下载
  - `components/FileList.tsx`：列出/刷新文件，`GET /api/files`；支持下载、校验（`/files/{id}/sha256`）与删除
  - `components/Chat.tsx`：WS 连接与消息收发，显示连接状态
  - `components/Devices.tsx`：设备列表 `GET /api/devices`
  - `lib/config.ts`：`useApiConfig()` 从 `/api/config` 解析后端地址与端点
  - `lib/auth.ts`：`useAuth()` 前端 token 管理（localStorage 持久化）
  - `app/layout.tsx`：全局布局，注入 Inter 字体与 PWA manifest
  - `app/page.tsx`：主页面（上传/下载/聊天/设备 Tab + 状态栏 + 配对）

- 交互流程
  - 配置加载：页面启动 `GET /api/config` 获取 `api_base`、`upload`、`ws` 等端点
  - 设备配对：用户输入一次性令牌 → `POST /api/pair` → 前端保存返回的 token（JWT）
  - 上传文件：选择文件 → `tus.Upload(endpoint: "/tus/files")` → 创建会话/分块上传 → 服务端完成后计算 SHA-256 与元数据 → 前端显示校验/下载入口
  - 下载与校验：`GET /files/{id}` 支持 Range；`GET /files/{id}/sha256` 获取校验和
  - 消息通信：前端建立 `ws://.../ws` 连接（带 token），发送与接收消息

- 开发代理（Next rewrites）
  - 在开发模式，Next 将同源请求代理到后端 `http://localhost:3280`，避免 CORS 与 JSON 解析错误：

```ts
// web/client/next.config.ts
export default {
  async rewrites() {
    return [
      { source: "/api/:path*",  destination: "http://localhost:3280/api/:path*" },
      { source: "/files/:path*", destination: "http://localhost:3280/files/:path*" },
      { source: "/upload/:path*", destination: "http://localhost:3280/upload/:path*" },
      { source: "/tus/:path*",    destination: "http://localhost:3280/tus/:path*" },
    ];
  },
};
```

- 架构示意（开发态）
```
浏览器 (http://localhost:3000)
      ↓ 同源请求（fetch/WS/tus）
Next 开发服务器
      ├─ /api/*    → http://localhost:3280/api/*
      ├─ /files/*  → http://localhost:3280/files/*
      └─ /tus/*    → http://localhost:3280/tus/*
后端 Go 服务器 (http://localhost:3280)
```

---

## 项目结构
```
easy-sync/
├── cmd/server/          # 后端入口
├── pkg/                 # 核心包
│   ├── config/          # 配置管理
│   ├── server/          # HTTP 服务器（Gin）
│   ├── websocket/       # WebSocket 处理
│   ├── upload/          # 文件上传 (tusd handler, /tus/*)
│   ├── download/        # 文件下载与校验
│   ├── discovery/       # mDNS/Bonjour 服务发现
│   └── security/        # 认证与配对
├── web/
│   ├── client/          # Next.js 前端工程
│   │   ├── src/app      # 布局与页面
│   │   ├── src/components  # 组件
│   │   ├── src/lib      # 配置/鉴权工具
│   │   └── public       # 前端静态资源（manifest 等）
│   └── public/          # 旧版静态页（index.html, sw.js）
└── docs/                # 文档
```

---

## 环境要求
- 操作系统：Windows / macOS / Linux
- Go：1.21+（建议设置 `GOPROXY` 加速依赖）
- Node.js：18+（建议 20+），npm（或 pnpm/yarn）

---

## 运行步骤

- 后端（开发态）
```powershell
# Windows PowerShell
$env:GOPROXY = "https://goproxy.cn,direct"
# 安装依赖
go mod download

# 后端build
cd cmd/server; go build -o easy-sync
# 启动服务器（默认 3280）
go run ./cmd/server
# 控制台将打印：
# HTTP:   http://[::]:3280
# Health: http://[::]:3280/health
# API:    http://[::]:3280/api/config
```

- 前端（开发态）
```bash
cd web/client
npm install
npm run dev
# 浏览器访问 http://localhost:3000/
# 所有 /api/*、/files/*、/tus/* 请求自动代理到后端 3280
```

- 前端（生产态，示例）
```bash
# 构建并启动 Next 服务器
cd web/client
npm run build
npm run start  # 默认 3000 端口
# 将 3000 的请求通过反向代理（nginx/caddy）转发到后端 /api、/files、/tus、/ws
```

- Docker（仅后端）
```bash
docker build -t easy-sync .
docker run -p 3280:3280 -v $(pwd)/uploads:/app/uploads easy-sync
```

---

## 接口文档

- 配置与健康
  - `GET /health` - 健康检查
  - `GET /api/config` - 返回服务端地址与端点

  示例：
  ```bash
  curl http://localhost:3000/api/config
  # { "endpoints": { "api_base": "http://[::]:3280/api", "upload": "http://[::]:3280/tus/", "websocket": "ws://[::]:3280/ws" }, ... }
  ```

- 设备配对
  - `GET /api/qr` - 获取二维码/令牌（需认证）
  - `POST /api/pair` - 完成设备配对，返回前端 token（JWT）
  - `GET /api/devices` - 获取已配对设备列表（需认证）
  - `DELETE /api/devices/{id}` - 删除配对设备（需认证）

  示例（配对）：
  ```bash
  curl -X POST http://localhost:3000/api/pair \
    -H "Content-Type: application/json" \
    -d '{"token":"PAIR_TOKEN","device_id":"<uuid>","device_name":"Desktop Browser"}'
  # => { "token": "<JWT>" }
  ```

- 文件操作
  - `POST /tus/files` - 创建上传会话（TUS 协议）
  - `PATCH /tus/files/{id}` - 分块上传（TUS 协议）
  - `HEAD /tus/files/{id}` - 查询上传状态（TUS 协议）
  - `GET /files/{id}` - 下载（支持 Range）
  - `GET /files/{id}/sha256` - 获取校验和
  - `GET /api/files` - 文件列表（需认证）
  - `DELETE /api/files/{id}` - 删除（需认证）

- 消息通信
  - `GET /api/messages` - 获取历史消息（需认证）
  - `WS /ws` - WebSocket 实时通信

---

## 代码示例

- TUS 断点续传（前端）
```ts
import * as tus from "tus-js-client";

const upload = new tus.Upload(file, {
  endpoint: "/tus/files",
  retryDelays: [0, 1000, 3000, 5000],
  metadata: { filename: file.name, filetype: file.type },
  headers: { Authorization: `Bearer ${token}` },
  onProgress: (uploaded, total) => {
    const pct = ((uploaded / total) * 100).toFixed(2);
    console.log(`${file.name}: ${pct}%`);
  },
  onSuccess: () => {
    const id = (upload.url?.split("/").pop() || "").replace(/[^a-zA-Z0-9_-]/g, "");
    console.log(`上传完成: ${file.name} (id=${id})`);
  },
});
upload.start();
```

- WebSocket（前端，携带 token）
```ts
const url = new URL("ws://localhost:3280/ws");
url.searchParams.set("token", token); // useAuth() 获取
const ws = new WebSocket(url.toString());
ws.onopen = () => ws.send(JSON.stringify({ type: "chat", text: "hello" }));
ws.onmessage = (ev) => console.log("recv:", ev.data);
ws.onerror = (e) => console.error("ws error", e);
```

- 配对（前端）
```ts
await fetch("/api/pair", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ token, device_id: crypto.randomUUID(), device_name: "Desktop Browser" }),
});
```

- Next 开发代理（前端）
```ts
export default {
  async rewrites() {
    return [
      { source: "/api/:path*",  destination: "http://localhost:3280/api/:path*" },
      { source: "/files/:path*", destination: "http://localhost:3280/files/:path*" },
      { source: "/tus/:path*",   destination: "http://localhost:3280/tus/:path*" },
    ];
  },
};
```

---

## 常见问题与排查
- `Unexpected token '<'`：表示前端请求返回了 HTML（通常为 404 页面）而非 JSON。开发态请确保 `next.config.ts` 的 rewrites 生效，并通过 `http://localhost:3000/api/config` 验证返回 JSON。
- CORS 问题：开发态通过同源代理（rewrites）避免；生产态建议使用 Nginx/Caddy 做统一反向代理。
- WebSocket 地址：`/api/config` 返回为 `ws://[::]:3280/ws`，部分浏览器对 IPv6 显示不友好，建议在生产反代层统一到主机名或 `localhost`。

---

## 快速开始（摘要）
```bash
# 后端
go run ./cmd/server  # 端口 3280
# 前端
cd web/client && npm i && npm run dev  # 端口 3000
# 浏览器访问 http://localhost:3000/ 进行配对与传输
```

---

## 技术栈
- 后端：Go (Golang) + Gin + gorilla/websocket + tusd + logrus + zeroconf
- 前端：Next.js + Tailwind v4 + tus-js-client + Heroicons + PWA manifest

---

## claude code cli 使用体验

### 对中文支持不友好
- [中文字符将会乱码](https://github.com/anthropics/claude-code/issues/2780)

## 贡献与许可证
- 贡献流程：Fork → 功能分支 → 提交更改 → PR
- 许可证：MIT