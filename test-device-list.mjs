import { chromium } from '@playwright/test';

async function testDeviceList() {
  console.log('🚀 开始设备列表测试...\n');

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

      console.log(`📱 客户端 ${i + 1}: 正在打开网页...`);
      await page.goto('http://localhost:3003');
      await page.waitForTimeout(3000);
    }

    console.log('\n⏳ 等待所有客户端完成自动配对和WebSocket连接...\n');

    // 等待每个客户端连接成功
    for (let i = 0; i < pages.length; i++) {
      try {
        await pages[i].waitForSelector('text=已连接', { timeout: 15000 });
        console.log(`✅ 客户端 ${i + 1} 已连接`);
      } catch (e) {
        console.log(`❌ 客户端 ${i + 1} 连接失败`);
      }
    }

    // 切换到设备标签
    console.log('\n📱 切换所有客户端到设备页面...');
    for (let i = 0; i < pages.length; i++) {
      await pages[i].click('button:has-text("设备")');
      await pages[i].waitForTimeout(1000);
    }

    // 点击刷新按钮
    console.log('\n🔄 刷新设备列表...');
    await pages[0].click('button:has-text("刷新")');
    await pages[0].waitForTimeout(1000);

    // 验证设备列表
    console.log('\n✅ 验证设备列表:');
    for (let i = 0; i < pages.length; i++) {
      const deviceCount = await pages[i].locator('.space-y-2 > div').count();
      console.log(`\n客户端 ${i + 1} 看到的设备数量: ${deviceCount}`);

      if (deviceCount > 0) {
        const devices = await pages[i].locator('.space-y-2 > div').allTextContents();
        devices.forEach((device, idx) => {
          console.log(`  设备 ${idx + 1}: ${device.trim()}`);
        });
      } else {
        const emptyText = await pages[i].locator('text=暂无设备').textContent();
        console.log(`  ${emptyText}`);
      }
    }

    console.log('\n✨ 测试完成! 预期结果: 每个客户端应该看到3个设备');
    console.log('按 Ctrl+C 关闭浏览器');

    // 保持浏览器打开以便手动检查
    await new Promise(() => {});

  } catch (error) {
    console.error('❌ 测试过程中出错:', error);
  }
}

testDeviceList();
