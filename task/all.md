
# 目标与约束
- 纯局域网可用：无公网、不依赖第三方云服务
- 双向传输：手机→电脑上传，电脑→手机分发/下载
- 大文件支持：断点续传、校验、进度与速率显示
- 消息聊天：在线文本消息、已读与送达确认
- 易用：尽量手机无需装 App（优先浏览器/PWA），电脑端提供一键运行
- 安全：设备配对、局域网内 TLS 加密（可选），控制访问权限

# 推荐总体架构（MVP）
以“电脑端小型本地服务 + 手机端浏览器/PWA”为核心，先满足 80% 场景，再逐步扩展。

- 电脑端 Agent（Go/Rust 优先）
  - 启动本地 HTTP 服务，监听 0.0.0.0:\<port>
  - 提供 Web UI（单页应用 SPA，供手机/电脑浏览器访问）
  - 提供 API：文件上传下载、消息 WebSocket
  - mDNS/Bonjour 服务发现：广播 `_lanxfer._tcp`，名称包含设备名与端口
  - 首次配对：二维码 + 一次性 Token，建立受信会话
  - 可选 HTTPS：自签名证书，展示指纹，避免中间人风险
- 手机端
  - 直接用浏览器访问：`http://hostname.local:port` 或扫码 `http://<ip>:port?t=<token>`
  - 无需安装：可保存为 PWA（离线 UI、主屏图标）
  - 发送文件：通过页面上传（断点续传）
  - 接收文件：“拉取下载”为主（由电脑端推送文件链接/Offer，手机点击下载）

说明：手机“被动接收（自动保存）”在 iOS 上受限制（前台/用户交互要求），因此推荐以“拉取下载”实现可靠跨平台体验。若你确实需要“自动落地到相册/下载目录”，再追加移动端原生 App（Flutter/React Native 实现）。

# 功能模块设计
## 1) 发现与连接
- mDNS/Bonjour 广播服务 `_lanxfer._tcp`，文本记录包含：
  - device_name=“My-PC”
  - port=“3280”
  - supports=https|http
- Fallback：二维码与手动输入
  - UI 显示“扫码连接”，二维码内容为 URL：`http(s)://<host>:<port>?t=<one-time-token>`
  - 手机端可输入 IP:Port 直连（跨子网时使用）

## 2) 配对与安全
- 首次启动：
  - 电脑端生成 Ed25519 公私钥 + 自签名 TLS（可选）
  - 显示一次性配对 Token（或用二维码封装）
- 手机端访问时：
  - 携带 Token 完成一次性配对，服务器签发短期 Session（Set-Cookie + SameSite）
  - 建立“已配对设备”列表（设备指纹/公钥绑定），后续免 Token 登录
- 传输加密选型：
  - 局域网 HTTP：最快速，易用；配合一次性 Token/Session 也能较安全
  - 局域网 HTTPS：更安全，需在手机端信任证书或接受自签名风险提示
- 访问控制：
  - CORS 限制源域
  - CSRF 防护（SameSite + Token）
  - WebSocket 鉴权（Cookie/Token/子协议）

## 3) 文件传输
- 手机→电脑（上传）
  - 协议：HTTP(S) + 断点续传（推荐 tus 协议）
  - 流程：
    1) 手机发起创建：POST /tus/files 创建 upload resource，返回 upload URL
    2) 分块 PATCH 上传，失败可恢复
    3) 上传完成后服务端校验 SHA-256，落盘并生成文件记录
  - 优点：浏览器原生易用，断点续传生态成熟（tus-js-client、tusd/tusd-go）
- 电脑→手机（分发/下载）
  - 控制面：通过 WebSocket 发送 file_offer（包含文件名、大小、mime、下载 URL、过期时间）
  - 数据面：手机点击下载，GET /files/{id} 支持 Range 与多连接并发下载
  - 大文件优化：
    - sendfile/零拷贝、异步 IO
    - Range 并发分段下载（客户端可选）
    - 校验：提供 /files/{id}/sha256

- 断点续传与校验
  - 上传：tus 原生支持断点续传
  - 下载：通过 Range + ETag + If-Range 支持续传；提供 SHA-256 校验接口

- 目录与权限
  - 服务器端有“收件箱”目录
  - 可选“共享目录”，只读暴露，便于手机浏览与下载
  - 权限策略：已配对设备可访问；临时访客需一次性 Token

## 4) 消息聊天
- 通道：WebSocket `/ws`
- 消息类型（JSON）：
  - hello：握手与状态
  - chat：文本消息（支持 Markdown 简单渲染）
  - file_offer：文件可用通知（含下载链接、大小、mime、sha256）
  - delivery_ack：送达确认
  - typing/presence：对端正在输入/在线状态
- 示例：
```json
// client -> server
{ "type": "hello", "device": "iPhone-15-Pro", "capabilities": ["upload","download","chat"] }

{ "type": "chat", "id": "c-1731", "text": "测试一下", "timestamp": 1731672000 }

{ "type": "file_offer_ack", "offer_id": "f-991", "accepted": true }
```

```json
// server -> client
{ "type": "chat", "id": "c-1732", "from": "My-PC", "text": "已收到", "timestamp": 1731672003 }

{ "type": "file_offer",
  "offer_id": "f-991",
  "from": "My-PC",
  "name": "视频.mp4",
  "size": 2147483648,
  "mime": "video/mp4",
  "sha256": "…",
  "url": "https://<host>:<port>/files/f-991?sig=…&exp=…"
}
```

- 存储与同步：
  - 服务端保存会话与消息（轻量 SQLite/BadgerDB）
  - 历史消息分页拉取：`GET /api/messages?before=…&limit=…`

## 5) Web UI 交互
- 设备列表与状态：发现到的设备、连接状态、延迟、带宽估计
- 快速发送：
  - PC 端拖拽到页面/托盘图标，即发布 file_offer
  - 手机端选择文件后直接上传（可多选、显示队列）
- 聊天区：并列的消息/传输气泡、进度条、失败可重试
- 连接入口：扫码/手输 IP、最近设备、mDNS 发现
- 设置：配对管理、证书指纹、共享目录、速率限速、仅在 Wi‑Fi 传输等

## 6) 技术栈与关键库
- 后端（单文件/跨平台首选 Go；Rust 亦可）
  - HTTP/WS：Go net/http + gorilla/websocket 或 fasthttp + nhooyr/ws
  - Resumable Upload：tus-go-server（或自行实现基于 tus 协议）
  - mDNS/Bonjour：github.com/grandcat/zeroconf
  - 数据库：SQLite（modernc.org/sqlite）或嵌入式 KV（bbolt/badger）
  - TLS：自签名证书生成 + 指纹展示（SHA-256）
  - 传输优化：sendfile、Gzip/Br 压缩（文本类）、HTTP/2（或 QUIC/HTTP/3 作为增强）
- 前端
  - 框架：React/Vue/Svelte 任一
  - 上传：tus-js-client
  - WebSocket：浏览器原生
  - PWA：manifest + service worker（缓存 UI、离线可用）
- 原生移动端（可选增强）
  - Flutter/React Native，支持通知、前台保持连接、更好的文件权限

## 7) API 概要
- 文件
  - POST /tus/files           创建上传会话（tus）
  - PATCH /tus/files/{id}     分块上传（tus）
  - HEAD  /tus/files/{id}     查询上传偏移（tus）
  - GET   /files/{id}         下载（支持 Range）
  - GET   /files/{id}/sha256  获取校验和
  - GET   /api/files          列表与状态
  - DELETE/api/files/{id}     删除
- 消息
  - GET   /api/messages?before&limit
  - WS    /ws                 双向消息
- 配对与设备
  - GET   /api/qr             获取连接二维码
  - POST  /api/pair           提交一次性 Token 完成配对
  - GET   /api/devices        已配对设备列表
  - DELETE/api/devices/{id}   解除配对

## 8) 性能与可靠性
- 大文件（>2GB）：
  - 磁盘直通 + 零拷贝
  - Range + 并发下载
  - tus 断点续传 + 校验和
- 弱网与掉线：
  - WebSocket 自动重连
  - tus 恢复点查询 + 分块重试
  - 下载支持 If-Range 续传
- 并发与限速：
  - 限制每设备并发数/速率，避免占满路由器
  - 背压与队列优先级（聊天优先、文件次之）

## 9) 跨子网与企业网络
- mDNS 可能受限：提供“手动连接”模式，记忆常用设备
- 企业防火墙：支持自定义端口、仅 HTTP 模式，必要时单向（只有 PC 开端口）
- 可选中继：在同一内网但跨 VLAN 且路由不通时，提供“PC 作为中继”的内部转发

## 10) 快速实现路线图
- 里程碑 1（1–2 天）：最小可用
  - PC 端 HTTP + Web UI，手机可访问
  - 文件上传（简单 POST）、下载（GET）
  - 简单 WebSocket 聊天
- 里程碑 2（2–4 天）：可靠传输
  - 引入 tus 断点续传、进度、失败重试
  - Range 下载 + 校验和
  - file_offer/ack 模式与消息持久化
- 里程碑 3（3–5 天）：发现与安全
  - mDNS 广播与发现
  - 一次性配对 Token + 已配对设备
  - 自签名 HTTPS（可选）
- 里程碑 4（可选增强）
  - PWA、离线 UI、桌面托盘集成、系统分享菜单
  - HTTP/2/QUIC 性能优化、剪贴板同步、屏幕文本直发

## 11) 替代/增强方案（按需选）
- WebDAV 暴露共享目录
  - iOS“文件”App 原生支持 WebDAV，方便浏览/上传/下载
  - 可作为电脑→手机的“浏览式”拉取补充
- WebRTC DataChannel
  - 浏览器端 P2P 传输与拥塞控制优秀，但需信令
  - 本地内置信令（同一服务器）可行；纯离线时可用局域信令
- 原生 App
  - 若需手机“被动接收并自动保存”，需移动端常驻/前台权限

## 12) 测试清单
- 平台矩阵：iOS Safari/Chrome、Android Chrome、Windows/macOS/Linux 各主流浏览器
- 文件类型与体积：小文件、超大文件（>10GB）
- 网络情况：弱 Wi‑Fi、切换 AP、信号衰弱
- 断点续传：中断/重连恢复
- 安全：未配对拒绝访问、Token 过期、证书指纹校验
- mDNS：启用/禁用下的发现与回退流程

# 总结
该方案优先交付“电脑端单程序 + 手机浏览器即用”的最小可靠路径，用 tus 实现稳定断点续传，用 WebSocket 做控制与消息，用 mDNS 提升易用性，用一次性 Token 降低风险。后续按需引入 HTTPS、自签名证书、PWA、WebRTC、原生 App 等增强能力。
