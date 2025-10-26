import { test, expect } from '@playwright/test';
import {
  waitForConnection,
  sendChatMessage,
  getChatMessages,
  switchToChatTab,
  openPage,
  getConnectionStatus
} from './helpers';

/**
 * 并发连接压力测试
 */

test.describe('并发连接测试', () => {
  test('5个客户端同时连接', async ({ browser }) => {
    console.log('📱 测试5个客户端同时连接...');

    const clientCount = 5;

    // 创建5个独立的浏览器上下文
    const contexts = await Promise.all(
      Array.from({ length: clientCount }, () => browser.newContext())
    );

    const pages = await Promise.all(
      contexts.map(context => context.newPage())
    );

    try {
      // 同时打开所有页面
      console.log(`打开 ${clientCount} 个客户端...`);
      await Promise.all(pages.map(page => openPage(page)));

      // 等待所有客户端连接（增加超时时间）
      console.log('等待所有客户端连接...');
      const connectionResults = await Promise.allSettled(
        pages.map((page, i) =>
          waitForConnection(page, 20000).then(() => {
            console.log(`✅ 客户端 ${i + 1} 已连接`);
            return true;
          })
        )
      );

      // 统计成功连接的客户端数量
      const successCount = connectionResults.filter(
        r => r.status === 'fulfilled'
      ).length;

      console.log(`成功连接: ${successCount}/${clientCount}`);
      expect(successCount).toBe(clientCount);

      // 验证所有客户端的状态
      const statuses = await Promise.all(
        pages.map(page => getConnectionStatus(page))
      );

      statuses.forEach((status, i) => {
        console.log(`客户端 ${i + 1} 状态: ${status}`);
        expect(status).toBe('已连接');
      });

      console.log('✅ 所有客户端成功建立连接');
    } finally {
      await Promise.all(contexts.map(context => context.close()));
    }
  });

  test('并发消息发送和广播', async ({ browser }) => {
    console.log('📱 测试并发消息发送和广播...');

    const clientCount = 5;

    const contexts = await Promise.all(
      Array.from({ length: clientCount }, () => browser.newContext())
    );

    const pages = await Promise.all(
      contexts.map(context => context.newPage())
    );

    try {
      // 连接所有客户端
      await Promise.all(pages.map(page => openPage(page)));
      await Promise.all(pages.map(page => waitForConnection(page, 20000)));
      await Promise.all(pages.map(page => switchToChatTab(page)));

      console.log('所有客户端已连接到聊天页面');

      // 每个客户端发送一条消息
      console.log('开始并发发送消息...');
      await Promise.all(
        pages.map((page, i) =>
          sendChatMessage(page, `消息来自客户端 ${i + 1}`)
        )
      );

      // 等待消息传播
      await pages[0].waitForTimeout(2000);

      // 验证每个客户端都收到了一些消息
      console.log('\n验证消息接收:');
      for (let i = 0; i < pages.length; i++) {
        const messages = await getChatMessages(pages[i]);
        console.log(`客户端 ${i + 1} 收到 ${messages.length} 条消息`);
        expect(messages.length).toBeGreaterThan(0);
      }

      console.log('✅ 并发消息发送和广播成功');
    } finally {
      await Promise.all(contexts.map(context => context.close()));
    }
  });

  test('压力测试：快速连接和断开', async ({ browser }) => {
    console.log('📱 压力测试：快速连接和断开...');

    const rounds = 3;

    for (let round = 0; round < rounds; round++) {
      console.log(`\n第 ${round + 1} 轮测试...`);

      const contexts = await Promise.all([
        browser.newContext(),
        browser.newContext(),
        browser.newContext(),
      ]);

      const pages = await Promise.all(
        contexts.map(context => context.newPage())
      );

      // 连接
      await Promise.all(pages.map(page => openPage(page)));
      await Promise.all(pages.map(page => waitForConnection(page, 15000)));

      console.log(`  ✅ ${pages.length} 个客户端已连接`);

      // 断开
      await Promise.all(contexts.map(context => context.close()));
      console.log(`  ✅ ${pages.length} 个客户端已断开`);

      // 短暂等待
      await new Promise(resolve => setTimeout(resolve, 1000));
    }

    console.log(`\n✅ 完成 ${rounds} 轮快速连接/断开测试`);
  });
});
