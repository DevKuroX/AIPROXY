const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  const page = await context.newPage();

  page.on('console', msg => console.log('BROWSER LOG:', msg.text()));
  page.on('pageerror', error => console.log('PAGE ERROR:', error.message));
  page.on('response', response => {
    if (response.url().includes('/api/')) {
      console.log('API RESPONSE:', response.url(), response.status());
    }
  });
  page.on('requestfailed', request => {
    if (request.url().includes('/api/')) {
      console.log('API REQUEST FAILED:', request.url(), request.failure().errorText);
    }
  });

  try {
    console.log('Navigating to login page...');
    await page.goto('http://localhost:1433/login');
    
    console.log('Filling credentials...');
    await page.fill('input[type="text"]', 'admin');
    await page.fill('input[type="password"]', 'admin123');
    
    console.log('Submitting login...');
    await page.click('button[type="submit"]');
    
    console.log('Waiting for dashboard navigation...');
    await page.waitForURL('**/dashboard', { timeout: 5000 });
    
    console.log('Waiting for /api/me call...');
    await page.waitForTimeout(3000);
    
    console.log('Current URL:', page.url());
    
    const errorElement = await page.locator('text=/loading|error|failed/i').first().textContent().catch(() => null);
    if (errorElement) {
      console.log('Error/Loading text found:', errorElement);
    }
    
    await page.screenshot({ path: '/home/ubuntu/ai_proxy/tests/full-flow.png' });
    
  } catch (error) {
    console.error('Test failed:', error.message);
    await page.screenshot({ path: '/home/ubuntu/ai_proxy/tests/full-flow-error.png' });
  } finally {
    await browser.close();
  }
})();
