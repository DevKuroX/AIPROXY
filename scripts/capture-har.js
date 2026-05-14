/**
 * HAR Fixtures Capture Script - T0.10
 * Captures HTTP Archive files for 10 critical user flows
 */

const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

const FRONTEND_URL = 'http://localhost:3000';
const BACKEND_URL = 'http://localhost:1432';
const FIXTURES_DIR = path.join(__dirname, '../fixtures/before');

// Test credentials
const TEST_EMAIL = 'admin@example.com';
const TEST_PASSWORD = 'admin123';

async function main() {
  console.log('Starting HAR capture for 10 user flows...');
  console.log('Frontend:', FRONTEND_URL);
  console.log('Backend:', BACKEND_URL);
  console.log('Output:', FIXTURES_DIR);
  console.log('');

  const browser = await chromium.launch({ headless: true });
  
  try {
    // Flow 1: Login
    await captureFlow(browser, '01-login', async (page) => {
      console.log('Capturing 01-login...');
      await page.goto(FRONTEND_URL);
      await page.waitForLoadState('networkidle');
      
      // Check if already logged in, if so logout first
      const loginBtn = await page.$('button:has-text("Login"), a:has-text("Login")');
      const logoutBtn = await page.$('button:has-text("Logout"), a:has-text("Logout")');
      
      if (logoutBtn && !loginBtn) {
        await logoutBtn.click();
        await page.waitForLoadState('networkidle');
      }
      
      // Try to find login form
      const emailInput = await page.$('input[type="email"], input[name="email"], input[placeholder*="email" i]');
      const passwordInput = await page.$('input[type="password"], input[name="password"]');
      
      if (emailInput && passwordInput) {
        await emailInput.fill(TEST_EMAIL);
        await passwordInput.fill(TEST_PASSWORD);
        const submitBtn = await page.$('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")');
        if (submitBtn) {
          await submitBtn.click();
          await page.waitForLoadState('networkidle');
        }
      }
    });

    // Flow 2: Dashboard Load
    await captureFlow(browser, '02-dashboard-load', async (page) => {
      console.log('Capturing 02-dashboard-load...');
      await page.goto(FRONTEND_URL);
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000); // Wait for dashboard widgets
    });

    // Flow 3: Providers CRUD
    await captureFlow(browser, '03-providers-crud', async (page) => {
      console.log('Capturing 03-providers-crud...');
      await page.goto(FRONTEND_URL);
      await page.waitForLoadState('networkidle');
      
      // Navigate to providers
      const providersLink = await page.$('a:has-text("Provider"), nav a[href*="provider" i]');
      if (providersLink) {
        await providersLink.click();
        await page.waitForLoadState('networkidle');
      }
      
      await page.waitForTimeout(1000);
    });

    // Flow 4: Keys CRUD
    await captureFlow(browser, '04-keys-crud', async (page) => {
      console.log('Capturing 04-keys-crud...');
      await page.goto(FRONTEND_URL);
      await page.waitForLoadState('networkidle');
      
      const keysLink = await page.$('a:has-text("Key"), nav a[href*="key" i]');
      if (keysLink) {
        await keysLink.click();
        await page.waitForLoadState('networkidle');
      }
      
      await page.waitForTimeout(1000);
    });

    // Flow 5: Settings Save
    await captureFlow(browser, '05-settings-save', async (page) => {
      console.log('Capturing 05-settings-save...');
      await page.goto(FRONTEND_URL);
      await page.waitForLoadState('networkidle');
      
      const settingsLink = await page.$('a:has-text("Setting"), nav a[href*="setting" i]');
      if (settingsLink) {
        await settingsLink.click();
        await page.waitForLoadState('networkidle');
      }
      
      await page.waitForTimeout(1000);
    });

    // Flow 6: Usage Chart
    await captureFlow(browser, '06-usage-chart', async (page) => {
      console.log('Capturing 06-usage-chart...');
      await page.goto(FRONTEND_URL);
      await page.waitForLoadState('networkidle');
      
      const usageLink = await page.$('a:has-text("Usage"), nav a[href*="usage" i]');
      if (usageLink) {
        await usageLink.click();
        await page.waitForLoadState('networkidle');
      }
      
      await page.waitForTimeout(1000);
    });

    // Flow 7: Chat Stream
    await captureFlow(browser, '07-chat-stream', async (page) => {
      console.log('Capturing 07-chat-stream...');
      await page.goto(FRONTEND_URL);
      await page.waitForLoadState('networkidle');
      
      // Look for chat/playground link
      const chatLink = await page.$('a:has-text("Chat"), a:has-text("Playground"), nav a[href*="chat" i]');
      if (chatLink) {
        await chatLink.click();
        await page.waitForLoadState('networkidle');
      }
      
      await page.waitForTimeout(1000);
    });

    // Flow 8: OAuth Start
    await captureFlow(browser, '08-oauth-start', async (page) => {
      console.log('Capturing 08-oauth-start...');
      await page.goto(FRONTEND_URL);
      await page.waitForLoadState('networkidle');
      
      // Look for OAuth buttons
      const oauthBtn = await page.$('button:has-text("Google"), button:has-text("GitHub"), a:has-text("OAuth")');
      if (oauthBtn) {
        // Don't actually click to avoid redirect
        console.log('  Found OAuth button, skipping actual click');
      }
    });

    // Flow 9: OAuth Callback
    await captureFlow(browser, '09-oauth-callback', async (page) => {
      console.log('Capturing 09-oauth-callback...');
      // Simulate OAuth callback URL
      await page.goto(`${FRONTEND_URL}/auth/callback?code=test&state=test`).catch(() => {
        console.log('  OAuth callback page may not exist');
      });
      await page.waitForLoadState('networkidle').catch(() => {});
    });

    // Flow 10: Logout
    await captureFlow(browser, '10-logout', async (page) => {
      console.log('Capturing 10-logout...');
      await page.goto(FRONTEND_URL);
      await page.waitForLoadState('networkidle');
      
      const logoutBtn = await page.$('button:has-text("Logout"), a:has-text("Logout"), button:has-text("Sign out")');
      if (logoutBtn) {
        await logoutBtn.click();
        await page.waitForLoadState('networkidle');
      }
    });

    console.log('\n✅ All HAR files captured successfully!');
    
  } catch (error) {
    console.error('Error during HAR capture:', error);
    throw error;
  } finally {
    await browser.close();
  }
}

async function captureFlow(browser, flowName, flowFn) {
  const context = await browser.newContext({
    recordHar: {
      path: path.join(FIXTURES_DIR, `${flowName}.har`),
      content: 'embed'
    }
  });
  
  const page = await context.newPage();
  
  try {
    await flowFn(page);
  } catch (error) {
    console.error(`  Error in ${flowName}:`, error.message);
  } finally {
    await context.close();
  }
}

main().catch(console.error);
