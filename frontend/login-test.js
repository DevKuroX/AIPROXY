const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  const page = await context.newPage();

  page.on('console', msg => console.log('BROWSER LOG:', msg.text()));
  page.on('pageerror', error => console.log('PAGE ERROR:', error.message));
  page.on('requestfailed', request => console.log('REQUEST FAILED:', request.url(), request.failure().errorText));

  try {
    console.log('Navigating to login page...');
    await page.goto('http://localhost:1433/login', { waitUntil: 'networkidle', timeout: 10000 });
    
    console.log('Page loaded, taking screenshot...');
    await page.screenshot({ path: '/home/ubuntu/ai_proxy/tests/01-login-page.png' });
    
    console.log('Filling username...');
    await page.fill('input[type="text"]', 'admin');
    
    console.log('Filling password...');
    await page.fill('input[type="password"]', 'admin123');
    
    console.log('Taking screenshot before submit...');
    await page.screenshot({ path: '/home/ubuntu/ai_proxy/tests/02-before-submit.png' });
    
    console.log('Clicking submit button...');
    await page.click('button[type="submit"]');
    
    console.log('Waiting for navigation or error...');
    await page.waitForTimeout(3000);
    
    console.log('Taking screenshot after submit...');
    await page.screenshot({ path: '/home/ubuntu/ai_proxy/tests/03-after-submit.png' });
    
    console.log('Current URL:', page.url());
    console.log('Page title:', await page.title());
    
    const errorText = await page.locator('text=/error|failed|invalid/i').first().textContent().catch(() => null);
    if (errorText) {
      console.log('Error message found:', errorText);
    }
    
  } catch (error) {
    console.error('Test failed:', error.message);
    await page.screenshot({ path: '/home/ubuntu/ai_proxy/tests/error.png' });
  } finally {
    await browser.close();
  }
})();
