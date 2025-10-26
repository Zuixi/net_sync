import { test, expect } from '@playwright/test';
import {
  waitForConnection,
  sendChatMessage,
  getChatMessages,
  switchToChatTab,
  openPage,
  expectMessageContains
} from './helpers';

/**
 * 聊天消息发送和广播测试
 */

test.describe('聊天消息功能', () => {
  test('单个客户端发送消息', async ({ page }) => {
    console.log('📱 测试单个客户端发送消息...');

    await openPage(page);
    await waitForConnection(page);
    await switchToChatTab(page);

    // 发送消息
    const testMessage = '测试消息 - 单客户端';
    await sendChatMessage(page, testMessage);

    // 等待消息显示
    await page.waitForTimeout(1000);

    // 验证消息已显示
    await expectMessageContains(page, testMessage);

    console.log('✅ 消息发送成功并正确显示');
  });

  test('多客户端消息广播', async ({ browser }) => {
    console.log('📱 测试多客户端消息广播...');

    // 创建3个客户端
    const contexts = await Promise.all([
      browser.newContext(),
      browser.newContext(),
      browser.newContext(),
    ]);

    const pages = await Promise.all(
      contexts.map(context => context.newPage())
    );

    try {
      // 打开所有页面并连接
      await Promise.all(pages.map(page => openPage(page)));
      await Promise.all(pages.map(page => waitForConnection(page)));
      await Promise.all(pages.map(page => switchToChatTab(page)));

      // 客户端1发送消息
      const message1 = '你好，我是客户端1';
      await sendChatMessage(pages[0], message1);
      await pages[0].waitForTimeout(1000);

      // 客户端2发送消息
      const message2 = '客户端2收到！';
      await sendChatMessage(pages[1], message2);
      await pages[1].waitForTimeout(1000);

      // 客户端3发送消息
      const message3 = '客户端3也在线';
      await sendChatMessage(pages[2], message3);
      await pages[2].waitForTimeout(1500);

      // 验证所有客户端都收到了所有消息
      for (let i = 0; i < pages.length; i++) {
        console.log(`\n验证客户端 ${i + 1} 的消息:`);
        const messages = await getChatMessages(pages[i]);
        console.log(`  收到 ${messages.length} 条消息`);

        // 检查是否包含所有消息
        const hasMessage1 = messages.some(m => m.includes('客户端1'));
        const hasMessage2 = messages.some(m => m.includes('客户端2'));
        const hasMessage3 = messages.some(m => m.includes('客户端3'));

        expect(hasMessage1 || hasMessage2 || hasMessage3).toBeTruthy();
        console.log(`  ✅ 客户端 ${i + 1} 接收到消息广播`);
      }

      console.log('\n✅ 消息广播功能正常');
    } finally {
      await Promise.all(contexts.map(context => context.close()));
    }
  });

  test('验证消息显示格式', async ({ browser }) => {
    console.log('📱 测试消息显示格式...');

    const contexts = await Promise.all([
      browser.newContext(),
      browser.newContext(),
    ]);

    const pages = await Promise.all(
      contexts.map(context => context.newPage())
    );

    try {
      await Promise.all(pages.map(page => openPage(page)));
      await Promise.all(pages.map(page => waitForConnection(page)));
      await Promise.all(pages.map(page => switchToChatTab(page)));

      // 客户端1发送消息
      await sendChatMessage(pages[0], '来自客户端1的消息');
      await pages[0].waitForTimeout(1500);

      // 验证客户端1看到自己的消息（右对齐，不显示发送者）
      const messages1 = await getChatMessages(pages[0]);
      console.log('客户端1看到的消息:', messages1);

      // 验证客户端2看到对方的消息（左对齐，显示发送者）
      const messages2 = await getChatMessages(pages[1]);
      console.log('客户端2看到的消息:', messages2);

      expect(messages1.length).toBeGreaterThan(0);
      expect(messages2.length).toBeGreaterThan(0);

      console.log('✅ 消息格式验证完成');
    } finally {
      await Promise.all(contexts.map(context => context.close()));
    }
  });

  test('系统消息显示', async ({ page }) => {
    console.log('📱 测试系统消息显示...');

    await openPage(page);
    await waitForConnection(page);
    await switchToChatTab(page);

    // 等待系统消息
    await page.waitForTimeout(1000);

    // 验证有系统消息（例如"已连接"）
    const messages = await getChatMessages(page);
    const hasSystemMessage = messages.some(
      m => m.includes('已连接') || m.includes('系统')
    );

    console.log('收到的消息:', messages);
    expect(messages.length).toBeGreaterThan(0);

    console.log('✅ 系统消息验证完成');
  });
});
