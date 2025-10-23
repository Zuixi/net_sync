import { chromium } from '@playwright/test';

async function testMultipleClients() {
  console.log('🚀 开始多客户端测试...\n');

  const browser = await chromium.launch({ headless: false });
  const contexts = [];
  const pages = [];

  try {
    // 创建3个独立的浏览器上下文(模拟3个独立客户端)
    for (let i = 0; i < 3; i++) {
      const context = await browser.newContext();
      const page = await context.newPage();
      contexts.push(context);
      pages.push(page);

      // 监听浏览器console日志
      page.on('console', msg => {
        console.log(`[客户端${i+1} console.${msg.type()}]`, msg.text());
      });

      console.log(`📱 客户端 ${i + 1}: 正在打开网页...`);
      await page.goto('http://localhost:3003');

      // 等待页面加载
      await page.waitForTimeout(2000);
    }

    console.log('\n⏳ 等待所有客户端完成自动配对和WebSocket连接...\n');

    // 等待每个客户端连接成功(最多等待15秒)
    for (let i = 0; i < pages.length; i++) {
      try {
        await pages[i].waitForSelector('text=已连接', { timeout: 15000 });
        console.log(`✅ 客户端 ${i + 1} 已连接`);
      } catch (e) {
        const statusText = await pages[i].locator('text=/已连接|连接中|连接断开|未连接/').textContent();
        console.log(`❌ 客户端 ${i + 1} 连接失败,当前状态: ${statusText}`);
      }
    }

    // 切换到聊天标签
    console.log('\n💬 切换所有客户端到聊天页面...');
    for (let i = 0; i < pages.length; i++) {
      await pages[i].click('button:has-text("聊天")');
      await pages[i].waitForTimeout(1000);
    }

    // 从不同客户端发送消息
    console.log('\n📤 从客户端1发送消息: "你好,我是客户端1"');
    await pages[0].fill('input[placeholder="输入消息"]', '你好,我是客户端1');
    await pages[0].click('button:has-text("发送")');
    await pages[0].waitForTimeout(1000);

    console.log('📤 从客户端2发送消息: "客户端2收到!"');
    await pages[1].fill('input[placeholder="输入消息"]', '客户端2收到!');
    await pages[1].click('button:has-text("发送")');
    await pages[1].waitForTimeout(1000);

    console.log('📤 从客户端3发送消息: "客户端3也在线"');
    await pages[2].fill('input[placeholder="输入消息"]', '客户端3也在线');
    await pages[2].click('button:has-text("发送")');
    await pages[2].waitForTimeout(2000);

    // 验证广播功能 - 检查每个客户端是否都收到了所有消息
    console.log('\n✅ 验证广播功能:');
    for (let i = 0; i < pages.length; i++) {
      const messages = await pages[i].locator('.space-y-2 > div').allTextContents();
      console.log(`\n客户端 ${i + 1} 看到的消息列表 (共${messages.length}条):`);
      messages.forEach((msg, idx) => {
        console.log(`  ${idx + 1}. ${msg.trim()}`);
      });
    }

    console.log('\n✨ 测试完成! 请检查浏览器窗口和上面的日志');
    console.log('按 Ctrl+C 关闭浏览器');

    // 保持浏览器打开以便手动检查
    await new Promise(() => {});

  } catch (error) {
    console.error('❌ 测试过程中出错:', error);
  } finally {
    // 不自动关闭,方便观察
    // await browser.close();
  }
}

testMultipleClients();
