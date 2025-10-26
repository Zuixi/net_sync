import { defineConfig, devices } from '@playwright/test';

/**
 * Playwright 配置文件
 * 用于端到端测试 easy_sync 的后端和前端功能
 */
export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',

  use: {
    baseURL: 'http://localhost:3002',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  // 不启动 webServer，假设后端服务器已经在运行
  // 如需自动启动，可以添加 webServer 配置
});
