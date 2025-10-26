/**
 * 客户端配置模块
 * 统一管理所有环境变量和配置值
 */

// ============================================
// 辅助函数
// ============================================

/**
 * 从环境变量读取字符串值,提供默认值
 */
function getEnvString(key: string, defaultValue: string): string {
  return process.env[key] || defaultValue;
}

/**
 * 从环境变量读取数字值,提供默认值
 */
function getEnvNumber(key: string, defaultValue: number): number {
  const value = process.env[key];
  if (!value) return defaultValue;
  const parsed = parseInt(value, 10);
  return isNaN(parsed) ? defaultValue : parsed;
}

/**
 * 从环境变量读取逗号分隔的数字数组
 */
function getEnvNumberArray(key: string, defaultValue: number[]): number[] {
  const value = process.env[key];
  if (!value) return defaultValue;
  const parsed = value.split(',').map(v => parseInt(v.trim(), 10)).filter(v => !isNaN(v));
  return parsed.length > 0 ? parsed : defaultValue;
}

// ============================================
// API 配置
// ============================================

export const API_CONFIG = {
  /** Go 服务器基础 URL */
  BASE_URL: getEnvString('NEXT_PUBLIC_API_BASE_URL', 'http://localhost:3280'),

  /** WebSocket 端口 */
  WS_PORT: getEnvNumber('NEXT_PUBLIC_WS_PORT', 3280),
} as const;

// ============================================
// WebSocket 配置
// ============================================

export const WEBSOCKET_CONFIG = {
  /** 重连延迟 (毫秒) */
  RECONNECT_DELAY: getEnvNumber('NEXT_PUBLIC_WS_RECONNECT_DELAY', 3000),

  /** 消息历史记录限制 (条数) */
  MESSAGE_HISTORY_LIMIT: getEnvNumber('NEXT_PUBLIC_WS_MESSAGE_HISTORY_LIMIT', 200),
} as const;

// ============================================
// 文件上传配置
// ============================================

export const UPLOAD_CONFIG = {
  /** TUS 上传端点 */
  TUS_ENDPOINT: getEnvString('NEXT_PUBLIC_TUS_ENDPOINT', '/tus/files'),

  /** TUS 上传重试延迟 (毫秒数组) */
  TUS_RETRY_DELAYS: getEnvNumberArray('NEXT_PUBLIC_TUS_RETRY_DELAYS', [0, 1000, 3000, 5000]),

  /** 上传日志最大保留条数 */
  LOG_LIMIT: getEnvNumber('NEXT_PUBLIC_UPLOAD_LOG_LIMIT', 10),
} as const;

// ============================================
// 本地存储配置
// ============================================

export const STORAGE_CONFIG = {
  /** Token 在 localStorage 中的存储键名 */
  TOKEN_KEY: getEnvString('NEXT_PUBLIC_TOKEN_STORAGE_KEY', 'easy_sync_token'),

  /** Device ID 在 localStorage 中的存储键名 */
  DEVICE_ID_KEY: getEnvString('NEXT_PUBLIC_DEVICE_ID_STORAGE_KEY', 'easy_sync_device_id'),

  /** Device Name 在 localStorage 中的存储键名 */
  DEVICE_NAME_KEY: getEnvString('NEXT_PUBLIC_DEVICE_NAME_STORAGE_KEY', 'easy_sync_device_name'),
} as const;

// ============================================
// 设备命名配置
// ============================================

export const DEVICE_CONFIG = {
  /** 移动设备默认名称 */
  MOBILE_NAME: getEnvString('NEXT_PUBLIC_MOBILE_DEVICE_NAME', '移动设备'),

  /** 浏览器设备默认名称 */
  BROWSER_NAME: getEnvString('NEXT_PUBLIC_BROWSER_DEVICE_NAME', '网页浏览器'),

  /** 桌面浏览器默认名称 */
  DESKTOP_NAME: getEnvString('NEXT_PUBLIC_DESKTOP_DEVICE_NAME', 'Desktop Browser'),
} as const;

// ============================================
// 开发环境配置
// ============================================

export const DEV_CONFIG = {
  /** 开发环境允许的跨域源 */
  ALLOWED_ORIGIN: getEnvString('NEXT_PUBLIC_DEV_ALLOWED_ORIGIN', ''),
} as const;

// ============================================
// PWA 配置
// ============================================

export const PWA_CONFIG = {
  /** PWA 缓存名称版本 */
  CACHE_NAME: getEnvString('NEXT_PUBLIC_PWA_CACHE_NAME', 'easy-sync-v1'),
} as const;

// ============================================
// 导出所有配置的联合类型
// ============================================

export const CONFIG = {
  api: API_CONFIG,
  websocket: WEBSOCKET_CONFIG,
  upload: UPLOAD_CONFIG,
  storage: STORAGE_CONFIG,
  device: DEVICE_CONFIG,
  dev: DEV_CONFIG,
  pwa: PWA_CONFIG,
} as const;

// ============================================
// 类型导出
// ============================================

export type Config = typeof CONFIG;

// ============================================
// 默认导出
// ============================================

export default CONFIG;

// ============================================
// 辅助工具函数
// ============================================

/**
 * 获取完整的 WebSocket URL
 * @param path WebSocket 路径 (如 '/ws')
 * @returns 完整的 WebSocket URL
 */
export function getWebSocketURL(path: string = '/ws'): string {
  if (typeof window === 'undefined') {
    // 服务端渲染时返回默认值
    return `ws://localhost:${API_CONFIG.WS_PORT}${path}`;
  }

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.hostname;
  const port = API_CONFIG.WS_PORT;

  return `${protocol}//${host}:${port}${path}`;
}

/**
 * 获取完整的 API URL
 * @param path API 路径
 * @returns 完整的 API URL
 */
export function getAPIURL(path: string): string {
  // 如果路径已经是完整URL，直接返回
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path;
  }

  // 确保路径以 / 开头
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;

  return `${API_CONFIG.BASE_URL}${normalizedPath}`;
}

/**
 * 判断是否为开发环境
 */
export function isDevelopment(): boolean {
  return process.env.NODE_ENV === 'development';
}

/**
 * 判断是否为生产环境
 */
export function isProduction(): boolean {
  return process.env.NODE_ENV === 'production';
}

// ============================================
// React Hook
// ============================================

/**
 * 获取 API 配置的 React Hook
 * 注意：配置是静态的，这个 hook 主要用于类型一致性
 */
export function useApiConfig() {
  return {
    config: CONFIG,
    loading: false,
    error: null
  };
}