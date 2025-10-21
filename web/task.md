# 实现

## 目录结构
- 封装配置与鉴权：
  - src/lib/config.ts 提供 useApiConfig 从 /api/config 解析服务端地址与端点。
  - src/lib/auth.ts 提供 useAuth 的 token 持久化与清理。
- 组件化迁移现有功能：
  - StatusBar 展示连接状态。
  - Pairing 支持输入 token、调用 /api/pair 并保存返回 token。
  - Uploader 基于 tus-js-client 上传文件，显示进度与速率。
  - FileList 列出与刷新文件，支持校验（SHA-256）、下载与删除。
  - Chat 建立 WebSocket 连接、发送与接收消息，显示连接状态。
  - Devices 获取并展示设备列表。
- 布局与页面：
  - src/app/layout.tsx 添加 Inter 字体、暗色背景、注入 PWA manifest.json 。
  - src/app/page.tsx 顶部状态栏与功能 Tab（上传、下载、聊天、设备），底部版权与轻导航。
- 运行与预览：
  - 启动了开发服务器，日志显示地址 http://localhost:3000 。已打开预览并未见错误。

## 运行
如何运行

- 在`easy_sync\web\client` 执行 `npm run dev` 。
- 访问 `http://localhost:3000/` 查看页面与组件。

## 待优化
- 生产环境反代：在 Nginx/Caddy 层统一反代 /api 、 /files 、 /tus 、 /ws ，并将 /api/config 的返回标准化为域名，以避免 IPv6 地址带来的显示与兼容性问题。
- 错误通知：在前端统一错误提示（如 Toast），包括 JSON 解析错误、WS断线重连、上传失败重试等。
- E2E 验证：引入简单的测试（如 Playwright）验证上传/下载/配对/WS消息的完整流程。
- WebSocket：当前后端返回 ws://[::]:3280/ws ，部分浏览器可能不接受该 IPv6 地址格式。前端可在创建连接时替换为 ws://localhost:3280/ws ，或在后端配置里改为具体主机名；如需统一同源代理，建议用 Nginx/Caddy 在生产环境做 WS 反向代理。
- 生产配置：若不使用 dev 代理，前端可改为从环境变量读取后端地址（如 NEXT_PUBLIC_API_ORIGIN ），在 useApiConfig 将绝对地址规范化为同源相对路径。
- 在 Uploader 增加断点续传缓存（ tus-js-client 的 fingerprint + resume ）以提升大文件体验。
- 将 WebSocket 的消息模型与错误处理抽象为 src/lib/ws.ts ，便于复用与重连策略优化。
- 统一错误提示与加载态为低侵入通知组件（如 Toast），提升交互一致性。
- 在 FileList 的校验过程增加流式 hash 计算与进度反馈，避免大文件卡顿。