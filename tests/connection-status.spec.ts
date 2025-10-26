import { test, expect } from '@playwright/test';
import {
  waitForConnection,
  getConnectionStatus,
  openPage
} from './helpers';

/**
 * 连接状态同步测试
 */

test.describe('连接状态同步', () => {
  test('单个客户端连接状态正确显示', async ({ page }) => {
    console.log('📱 测试单个客户端连接状态...');

    await openPage(page);

    // 等待自动配对和连接
    await waitForConnection(page);

    // 验证状态为"已连接"
    const status = await getConnectionStatus(page);
    expect(status).toBe('已连接');

    console.log('✅ 客户端状态正确显示为: 已连接');
  });

  test('多个客户端独立连接状态', async ({ browser }) => {
    console.log('📱 测试多个客户端独立连接状态...');

    // 创建3个独立的浏览器上下文
    const contexts = await Promise.all([
      browser.newContext(),
      browser.newContext(),
      browser.newContext(),
    ]);

    const pages = await Promise.all(
      contexts.map(context => context.newPage())
    );

    try {
      // 打开所有页面
      await Promise.all(pages.map(page => openPage(page)));

      // 等待所有客户端连接
      await Promise.all(pages.map(page => waitForConnection(page)));

      // 验证所有客户端状态都是"已连接"
      const statuses = await Promise.all(
        pages.map(page => getConnectionStatus(page))
      );

      statuses.forEach((status, i) => {
        expect(status).toBe('已连接');
        console.log(`✅ 客户端 ${i + 1} 状态: ${status}`);
      });

      console.log('✅ 所有客户端独立连接成功');
    } finally {
      // 清理资源
      await Promise.all(contexts.map(context => context.close()));
    }
  });

  test('验证连接状态变化流程', async ({ page }) => {
    console.log('📱 测试连接状态变化流程...');

    await page.goto('http://localhost:3002');

    // 可能初始状态是"未连接"或"连接中"
    await page.waitForTimeout(500);
    const initialStatus = await getConnectionStatus(page);
    console.log(`初始状态: ${initialStatus}`);

    // 最终应该变为"已连接"
    await waitForConnection(page, 15000);
    const finalStatus = await getConnectionStatus(page);
    expect(finalStatus).toBe('已连接');

    console.log(`✅ 状态变化: ${initialStatus} -> ${finalStatus}`);
  });
});
