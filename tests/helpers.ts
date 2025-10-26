import { Page, expect } from '@playwright/test';

/**
 * 测试辅助函数
 */

/**
 * 等待客户端连接成功
 * @param page Playwright 页面对象
 * @param timeout 超时时间（毫秒）
 */
export async function waitForConnection(page: Page, timeout = 15000) {
  await page.waitForSelector('text=已连接', { timeout });
}

/**
 * 检查连接状态
 * @param page Playwright 页面对象
 * @returns 连接状态文本
 */
export async function getConnectionStatus(page: Page): Promise<string> {
  const statusElement = await page.locator('text=/已连接|连接中|连接断开|未连接/');
  return await statusElement.textContent() || '';
}

/**
 * 发送聊天消息
 * @param page Playwright 页面对象
 * @param message 消息内容
 */
export async function sendChatMessage(page: Page, message: string) {
  await page.fill('input[placeholder="输入消息"]', message);
  await page.click('button:has-text("发送")');
}

/**
 * 获取所有聊天消息
 * @param page Playwright 页面对象
 * @returns 消息文本数组
 */
export async function getChatMessages(page: Page): Promise<string[]> {
  const messages = await page.locator('.space-y-2 > div').allTextContents();
  return messages.map(m => m.trim());
}

/**
 * 切换到聊天标签页
 * @param page Playwright 页面对象
 */
export async function switchToChatTab(page: Page) {
  await page.click('button:has-text("聊天")');
  await page.waitForTimeout(500);
}

/**
 * 打开页面并等待加载
 * @param page Playwright 页面对象
 * @param url 页面 URL
 */
export async function openPage(page: Page, url = 'http://localhost:3002') {
  await page.goto(url);
  await page.waitForTimeout(2000);
}

/**
 * 验证消息是否包含特定文本
 * @param page Playwright 页面对象
 * @param text 要查找的文本
 */
export async function expectMessageContains(page: Page, text: string) {
  const messages = await getChatMessages(page);
  const found = messages.some(msg => msg.includes(text));
  expect(found).toBeTruthy();
}
