const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  const page = await context.newPage();

  page.on('console', msg => console.log('BROWSER LOG:', msg.text()));
  page.on('pageerror', error => console.log('PAGE ERROR:', error.message));
  page.on('requestfailed', request => console.log('REQUEST FAILED:', request.url(), request.failure().errorText));

  try {
    console.log('Setting token in localStorage...');
    await page.goto('http://localhost:1433/login');
    
    await page.evaluate(() => {
      localStorage.setItem('token', 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSIsImlzX2FkbWluIjp0cnVlLCJleHAiOjE3Nzg2ODAyNjQsImlhdCI6MTc3ODU5Mzg2NH0.rAe6-Y9DjUWf4J0XgHfZ2SKN-KqrOabrjjwBIrufujw');
    });
    
    console.log('Navigating to dashboard...');
    await page.goto('http://localhost:1433/dashboard', { waitUntil: 'networkidle', timeout: 10000 });
    
    console.log('Dashboard loaded, waiting 2 seconds...');
    await page.waitForTimeout(2000);
    
    console.log('Taking screenshot...');
    await page.screenshot({ path: '/home/ubuntu/ai_proxy/tests/dashboard.png' });
    
    console.log('Current URL:', page.url());
    console.log('Page title:', await page.title());
    
    const bodyText = await page.locator('body').textContent();
    console.log('Page content preview:', bodyText.substring(0, 500));
    
  } catch (error) {
    console.error('Test failed:', error.message);
    await page.screenshot({ path: '/home/ubuntu/ai_proxy/tests/dashboard-error.png' });
  } finally {
    await browser.close();
  }
})();
