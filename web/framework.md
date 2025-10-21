# Web Design

## Framework
- **Next.js**: Server-side rendering and static site generation
- **File Upload**: tus-js-client for resumable uploads
- **Real-time Communication**: Browser native WebSocket API
- **PWA**: Web app manifest + service worker for offline capabilities and home screen installation
- **CSS Framework**: Tailwind v4 CSS for utility-first styling
- **Icons**: Heroicons for scalable vector icons
- **Font**: Inter font family for readability
- **Color Scheme**: Dark mode support with Tailwind v4 color palette
- **Responsive Design**: Mobile-first approach with Tailwind v4 breakpoints
- **Accessibility**: ARIA roles and keyboard navigation
- **Performance**: sendfile, Gzip/Brotli compression, HTTP/2 support
- **Caching**: Browser cache headers for static assets
- **Page Style**:
  - technical style for server-side rendered pages
  - clean and minimalistic design for file uploads and chat
  - blur effect for background elements
  - rounded corners for chat messages
  - shadow for chat bubbles
  - hover effects for interactive elements
  - transition effects for smooth interactions
  - smooth scrolling for navigation
  - fixed header for easy access to controls
  - fixed footer for chat controls
  - scrollbar styling for chat area

## Components
- **File Uploader**: Customizable tus-js-client integration
- **Chat Interface**: Real-time messaging with WebSocket
- **Device List**: Dynamic display of connected devices
- **File List**: Show active uploads and downloads
- **Status Indicators**: Connection status, upload progress, and error messages

## Functionality
- **Device Discovery**: mDNS and manual IP entry for service discovery
- **File Sharing**: Secure file transfers with SHA-256 verification
- **Chat Messaging**: Text-based communication with device-to-device encryption
- **File Management**: Create, delete, and list files on the server
- **Performance Monitoring**: Real-time upload/download speed and device stats
- **Security**: TLS encryption for all communications, with self-signed certificate generation