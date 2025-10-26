import { test, expect } from '@playwright/test';
import {
  waitForConnection,
  sendChatMessage,
  switchToChatTab,
  openPage,
  getConnectionStatus
} from './helpers';

/**
 * 断线重连测试
 */

test.describe('断线重连', () => {
  test('模拟WebSocket断开和重连', async ({ page }) => {
    console.log('📱 测试WebSocket断开和重连...');

    await openPage(page);
    await waitForConnection(page);

    console.log('✅ 初始连接成功');

    // 在浏览器中执行脚本强制关闭WebSocket
    await page.evaluate(() => {
      // 查找并关闭所有WebSocket连接
      const wsContext = (window as any).__wsContext;
      if (wsContext && wsContext.close) {
        wsContext.close();
      }
    });

    console.log('⚠️ 已手动关闭WebSocket连接');

    // 等待状态变为"连接断开"
    await page.waitForTimeout(2000);
    const disconnectedStatus = await getConnectionStatus(page);
    console.log(`当前状态: ${disconnectedStatus}`);

    // 等待自动重连（WEBSOCKET_CONFIG.RECONNECT_DELAY = 3000ms）
    console.log('等待自动重连...');
    try {
      await waitForConnection(page, 15000);
      const reconnectedStatus = await getConnectionStatus(page);
      console.log(`重连后状态: ${reconnectedStatus}`);
      expect(reconnectedStatus).toBe('已连接');
      console.log('✅ 自动重连成功');
    } catch (error) {
      console.log('❌ 自动重连失败或超时');
      throw error;
    }
  });

  test('断线后功能恢复', async ({ browser }) => {
    console.log('📱 测试断线后功能恢复...');

    const contexts = await Promise.all([
      browser.newContext(),
      browser.newContext(),
    ]);

    const pages = await Promise.all(
      contexts.map(context => context.newPage())
    );

    try {
      // 连接两个客户端
      await Promise.all(pages.map(page => openPage(page)));
      await Promise.all(pages.map(page => waitForConnection(page)));
      await Promise.all(pages.map(page => switchToChatTab(page)));

      console.log('✅ 两个客户端已连接');

      // 客户端1发送消息
      await sendChatMessage(pages[0], '断线前的消息');
      await pages[0].waitForTimeout(1000);

      // 模拟客户端2断线
      await pages[1].evaluate(() => {
        const wsContext = (window as any).__wsContext;
        if (wsContext && wsContext.close) {
          wsContext.close();
        }
      });

      console.log('⚠️ 客户端2已断线');
      await pages[1].waitForTimeout(2000);

      // 等待客户端2重连
      console.log('等待客户端2重连...');
      try {
        await waitForConnection(pages[1], 15000);
        console.log('✅ 客户端2重连成功');

        // 测试重连后功能是否正常
        await sendChatMessage(pages[0], '重连后的消息');
        await pages[0].waitForTimeout(1500);

        console.log('✅ 断线后功能恢复测试完成');
      } catch (error) {
        console.log('❌ 客户端2重连失败');
        throw error;
      }
    } finally {
      await Promise.all(contexts.map(context => context.close()));
    }
  });

  test('多次断线重连稳定性', async ({ page }) => {
    console.log('📱 测试多次断线重连稳定性...');

    await openPage(page);
    await waitForConnection(page);

    const reconnectRounds = 3;

    for (let i = 0; i < reconnectRounds; i++) {
      console.log(`\n第 ${i + 1} 次断线重连测试...`);

      // 强制断开
      await page.evaluate(() => {
        const wsContext = (window as any).__wsContext;
        if (wsContext && wsContext.close) {
          wsContext.close();
        }
      });

      await page.waitForTimeout(2000);
      console.log('  ⚠️ 已断开连接');

      // 等待重连
      try {
        await waitForConnection(page, 15000);
        console.log('  ✅ 重连成功');
      } catch (error) {
        console.log(`  ❌ 第 ${i + 1} 次重连失败`);
        throw error;
      }

      await page.waitForTimeout(1000);
    }

    console.log(`\n✅ 完成 ${reconnectRounds} 次断线重连测试`);
  });

  test('检查心跳机制', async ({ page }) => {
    console.log('📱 测试心跳机制...');

    await openPage(page);
    await waitForConnection(page);

    console.log('✅ 初始连接成功');

    // 监听控制台日志查看ping/pong消息
    page.on('console', msg => {
      const text = msg.text();
      if (text.includes('ping') || text.includes('pong')) {
        console.log(`  心跳消息: ${text}`);
      }
    });

    // 保持连接一段时间以观察心跳
    console.log('保持连接60秒以观察心跳...');
    await page.waitForTimeout(60000);

    // 验证仍然保持连接
    const status = await getConnectionStatus(page);
    expect(status).toBe('已连接');

    console.log('✅ 心跳机制正常，连接保持稳定');
  });
});
