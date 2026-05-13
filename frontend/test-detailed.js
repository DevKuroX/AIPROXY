const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  const page = await context.newPage();

  const apiCalls = [];
  
  page.on('console', msg => console.log('BROWSER:', msg.text()));
  page.on('pageerror', error => console.log('ERROR:', error.message));
  
  page.on('request', request => {
    if (request.url().includes('/api/')) {
      console.log('→ REQUEST:', request.method(), request.url());
      console.log('  Headers:', JSON.stringify(request.headers(), null, 2));
    }
  });
  
  page.on('response', async response => {
    if (response.url().includes('/api/')) {
      const body = await response.text().catch(() => 'Could not read body');
      console.log('← RESPONSE:', response.status(), response.url());
      console.log('  Body:', body);
      apiCalls.push({ url: response.url(), status: response.status(), body });
    }
  });

  try {
    console.log('\n=== STEP 1: Login ===');
    await page.goto('http://localhost:1433/login');
    await page.fill('input[type="text"]', 'admin');
    await page.fill('input[type="password"]', 'admin123');
    await page.click('button[type="submit"]');
    
    console.log('\n=== STEP 2: Wait for redirect ===');
    await page.waitForTimeout(2000);
    console.log('Current URL:', page.url());
    
    console.log('\n=== STEP 3: Check localStorage ===');
    const token = await page.evaluate(() => localStorage.getItem('token'));
    console.log('Token in localStorage:', token ? token.substring(0, 50) + '...' : 'NULL');
    
    console.log('\n=== STEP 4: Check page state ===');
    const pageText = await page.locator('body').textContent();
    console.log('Page contains "Loading":', pageText.includes('Loading'));
    console.log('Page contains "Welcome":', pageText.includes('Welcome'));
    console.log('Page contains "admin":', pageText.includes('admin'));
    
    console.log('\n=== API Calls Summary ===');
    apiCalls.forEach((call, i) => {
      console.log(`${i+1}. ${call.url} → ${call.status}`);
    });
    
    await page.screenshot({ path: '/home/ubuntu/ai_proxy/tests/debug.png', fullPage: true });
    
  } catch (error) {
    console.error('Test failed:', error.message);
  } finally {
    await browser.close();
  }
})();
