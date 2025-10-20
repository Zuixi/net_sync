# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**easy_sync** is a local network file and message transfer tool designed for seamless communication between mobile devices and desktop computers. The project focuses on LAN-only transfers without requiring cloud services, with support for resumable uploads, PWA capabilities, and native mobile applications.

**Current Status**: Planning/design phase. The repository contains comprehensive documentation but no implemented source code yet.

## Technology Stack

### Backend (Go)
- **HTTP/WebSocket**: Go net/http + gorilla/websocket or fasthttp + nhooyr/ws
- **Resumable Upload**: tus-go-server (implements tus protocol for resumable file transfers)
- **Service Discovery**: github.com/grandcat/zeroconf for mDNS/Bonjour broadcasting
- **Database**: SQLite (modernc.org/sqlite) or embedded KV (bbolt/badger)
- **TLS**: Self-signed certificate generation with SHA-256 fingerprint display
- **Performance**: sendfile, Gzip/Brotli compression, HTTP/2 support

### Frontend (Vue.js)
- **Framework**: Vue.js for single-page application
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

## Development Commands (Expected)

### Backend (Go)
```bash
# Build the desktop agent
go build -o easy-sync ./cmd/server

# Run in development mode
go run ./cmd/server

# Run tests
go test ./...

# Build for different platforms
GOOS=windows GOARCH=amd64 go build -o easy-sync.exe ./cmd/server
GOOS=darwin GOARCH=arm64 go build -o easy-sync-mac ./cmd/server
GOOS=linux GOARCH=amd64 go build -o easy-sync-linux ./cmd/server
```

### Frontend (Vue.js)
```bash
# Install dependencies
npm install

# Development server
npm run dev

# Build for production
npm run build

# Run tests
npm run test

# PWA build
npm run build:pwa
```

### Testing
```bash
# End-to-end tests (when implemented)
npm run test:e2e

# Performance tests
npm run test:performance
```

## Key API Endpoints (Planned)

### File Operations
- `POST /tus/files` - Create upload session (tus protocol)
- `PATCH /tus/files/{id}` - Chunk upload (tus protocol)
- `HEAD /tus/files/{id}` - Query upload offset (tus protocol)
- `GET /files/{id}` - Download file (supports Range requests)
- `GET /files/{id}/sha256` - Get file checksum
- `GET /api/files` - List files with metadata
- `DELETE /api/files/{id}` - Delete file

### Messaging & Communication
- `GET /api/messages?before&limit` - Paginated message history
- `WS /ws` - WebSocket endpoint for real-time messaging

### Device Pairing & Discovery
- `GET /api/qr` - Get QR code for connection
- `POST /api/pair` - Complete pairing with one-time token
- `GET /api/devices` - List paired devices
- `DELETE /api/devices/{id}` - Remove device pairing

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

## Important Notes

- **LAN-only operation**: No cloud dependencies or internet connectivity required
- **Security-first approach**: All transfers require device pairing and authentication
- **Cross-platform focus**: Prioritize browser compatibility over native features initially
- **Progressive enhancement**: Start with basic HTTP transfers, add advanced features incrementally
- network request,should set http proxy first:
`$env:HTTP_PROXY="http://127.0.0.1:13824"`
`$env:HTTPS_PROXY="http://127.0.0.1:13824"`