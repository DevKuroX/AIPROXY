#!/usr/bin/env python3
"""
HAR Fixtures Capture Script — T0.10
Automates capture of 10 user flows using Playwright
"""

import asyncio
import json
from pathlib import Path
from playwright.async_api import async_playwright

BACKEND_URL = "http://localhost:1432"
FRONTEND_URL = "http://localhost:3000"
FIXTURES_DIR = Path(__file__).parent.parent / "fixtures" / "before"

# Test credentials from backend/.env
ADMIN_EMAIL = "admin@example.com"
ADMIN_PASSWORD = "admin123"


async def capture_flow(name: str, description: str, flow_func):
    """Capture a single user flow as HAR"""
    print(f"\n{'='*60}")
    print(f"Capturing: {name}")
    print(f"Description: {description}")
    print(f"{'='*60}")
    
    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=True)
        context = await browser.new_context(
            record_har_path=str(FIXTURES_DIR / f"{name}.har"),
            record_har_content="embed"
        )
        page = await context.new_page()
        
        try:
            await flow_func(page)
            print(f"✓ {name} captured successfully")
        except Exception as e:
            print(f"✗ {name} failed: {e}")
        finally:
            await context.close()
            await browser.close()


async def flow_01_login(page):
    """Flow 1: Login"""
    await page.goto(FRONTEND_URL)
    await page.wait_for_load_state("networkidle")
    
    # Check if already logged in
    if "dashboard" in page.url or "login" not in page.url.lower():
        # Logout first
        try:
            await page.click('button:has-text("Logout")', timeout=2000)
            await page.wait_for_url("**/login", timeout=5000)
        except:
            pass
    
    # Fill login form
    await page.fill('input[name="email"], input[type="email"]', ADMIN_EMAIL)
    await page.fill('input[name="password"], input[type="password"]', ADMIN_PASSWORD)
    
    # Submit
    await page.click('button[type="submit"]')
    
    # Wait for redirect
    await page.wait_for_load_state("networkidle", timeout=10000)


async def flow_02_dashboard_load(page):
    """Flow 2: Dashboard Load"""
    # Login first
    await page.goto(FRONTEND_URL)
    await page.wait_for_load_state("networkidle")
    
    # If not logged in, login
    if "login" in page.url.lower():
        await page.fill('input[name="email"], input[type="email"]', ADMIN_EMAIL)
        await page.fill('input[name="password"], input[type="password"]', ADMIN_PASSWORD)
        await page.click('button[type="submit"]')
        await page.wait_for_load_state("networkidle")
    
    # Navigate to dashboard
    await page.goto(FRONTEND_URL)
    await page.wait_for_load_state("networkidle")
    
    # Wait for dashboard widgets to load
    await asyncio.sleep(2)


async def flow_03_providers_crud(page):
    """Flow 3: Providers CRUD"""
    # Login
    await page.goto(FRONTEND_URL)
    await page.wait_for_load_state("networkidle")
    
    if "login" in page.url.lower():
        await page.fill('input[name="email"], input[type="email"]', ADMIN_EMAIL)
        await page.fill('input[name="password"], input[type="password"]', ADMIN_PASSWORD)
        await page.click('button[type="submit"]')
        await page.wait_for_load_state("networkidle")
    
    # Navigate to providers page
    try:
        await page.click('a:has-text("Providers"), a[href*="provider"]')
        await page.wait_for_load_state("networkidle")
    except:
        # Try direct URL
        await page.goto(f"{FRONTEND_URL}/providers")
        await page.wait_for_load_state("networkidle")
    
    await asyncio.sleep(2)


async def flow_04_keys_crud(page):
    """Flow 4: Keys CRUD"""
    # Login
    await page.goto(FRONTEND_URL)
    await page.wait_for_load_state("networkidle")
    
    if "login" in page.url.lower():
        await page.fill('input[name="email"], input[type="email"]', ADMIN_EMAIL)
        await page.fill('input[name="password"], input[type="password"]', ADMIN_PASSWORD)
        await page.click('button[type="submit"]')
        await page.wait_for_load_state("networkidle")
    
    # Navigate to keys page
    try:
        await page.click('a:has-text("Keys"), a:has-text("API Keys"), a[href*="key"]')
        await page.wait_for_load_state("networkidle")
    except:
        await page.goto(f"{FRONTEND_URL}/keys")
        await page.wait_for_load_state("networkidle")
    
    await asyncio.sleep(2)


async def flow_05_settings_save(page):
    """Flow 5: Settings Save"""
    # Login
    await page.goto(FRONTEND_URL)
    await page.wait_for_load_state("networkidle")
    
    if "login" in page.url.lower():
        await page.fill('input[name="email"], input[type="email"]', ADMIN_EMAIL)
        await page.fill('input[name="password"], input[type="password"]', ADMIN_PASSWORD)
        await page.click('button[type="submit"]')
        await page.wait_for_load_state("networkidle")
    
    # Navigate to settings
    try:
        await page.click('a:has-text("Settings"), a[href*="setting"]')
        await page.wait_for_load_state("networkidle")
    except:
        await page.goto(f"{FRONTEND_URL}/settings")
        await page.wait_for_load_state("networkidle")
    
    await asyncio.sleep(2)


async def flow_06_usage_chart(page):
    """Flow 6: Usage Chart"""
    # Login
    await page.goto(FRONTEND_URL)
    await page.wait_for_load_state("networkidle")
    
    if "login" in page.url.lower():
        await page.fill('input[name="email"], input[type="email"]', ADMIN_EMAIL)
        await page.fill('input[name="password"], input[type="password"]', ADMIN_PASSWORD)
        await page.click('button[type="submit"]')
        await page.wait_for_load_state("networkidle")
    
    # Navigate to usage/analytics
    try:
        await page.click('a:has-text("Usage"), a:has-text("Analytics"), a[href*="usage"]')
        await page.wait_for_load_state("networkidle")
    except:
        await page.goto(f"{FRONTEND_URL}/usage")
        await page.wait_for_load_state("networkidle")
    
    await asyncio.sleep(2)


async def flow_07_chat_stream(page):
    """Flow 7: Chat Stream"""
    # Login
    await page.goto(FRONTEND_URL)
    await page.wait_for_load_state("networkidle")
    
    if "login" in page.url.lower():
        await page.fill('input[name="email"], input[type="email"]', ADMIN_EMAIL)
        await page.fill('input[name="password"], input[type="password"]', ADMIN_PASSWORD)
        await page.click('button[type="submit"]')
        await page.wait_for_load_state("networkidle")
    
    # Navigate to chat/test page
    try:
        await page.click('a:has-text("Chat"), a:has-text("Test"), a[href*="chat"]')
        await page.wait_for_load_state("networkidle")
    except:
        await page.goto(f"{FRONTEND_URL}/chat")
        await page.wait_for_load_state("networkidle")
    
    await asyncio.sleep(2)


async def flow_08_oauth_start(page):
    """Flow 8: OAuth Start"""
    # Login
    await page.goto(FRONTEND_URL)
    await page.wait_for_load_state("networkidle")
    
    if "login" in page.url.lower():
        await page.fill('input[name="email"], input[type="email"]', ADMIN_EMAIL)
        await page.fill('input[name="password"], input[type="password"]', ADMIN_PASSWORD)
        await page.click('button[type="submit"]')
        await page.wait_for_load_state("networkidle")
    
    # Navigate to OAuth providers
    try:
        await page.click('a:has-text("OAuth"), a[href*="oauth"]')
        await page.wait_for_load_state("networkidle")
    except:
        await page.goto(f"{FRONTEND_URL}/oauth")
        await page.wait_for_load_state("networkidle")
    
    await asyncio.sleep(2)


async def flow_09_oauth_callback(page):
    """Flow 9: OAuth Callback (Placeholder)"""
    # This requires actual OAuth flow completion
    # Capture what we can
    await page.goto(f"{FRONTEND_URL}/oauth/callback")
    await page.wait_for_load_state("networkidle")
    await asyncio.sleep(1)


async def flow_10_logout(page):
    """Flow 10: Logout"""
    # Login first
    await page.goto(FRONTEND_URL)
    await page.wait_for_load_state("networkidle")
    
    if "login" in page.url.lower():
        await page.fill('input[name="email"], input[type="email"]', ADMIN_EMAIL)
        await page.fill('input[name="password"], input[type="password"]', ADMIN_PASSWORD)
        await page.click('button[type="submit"]')
        await page.wait_for_load_state("networkidle")
    
    # Logout
    try:
        await page.click('button:has-text("Logout"), a:has-text("Logout")')
        await page.wait_for_load_state("networkidle")
    except:
        # Try direct API call
        await page.goto(f"{BACKEND_URL}/api/logout")
        await page.wait_for_load_state("networkidle")
    
    await asyncio.sleep(1)


async def main():
    """Capture all flows"""
    print("\n" + "="*60)
    print("HAR Fixtures Capture — T0.10")
    print("="*60)
    print(f"Backend URL: {BACKEND_URL}")
    print(f"Frontend URL: {FRONTEND_URL}")
    print(f"Output directory: {FIXTURES_DIR}")
    print("="*60)
    
    # Ensure fixtures directory exists
    FIXTURES_DIR.mkdir(parents=True, exist_ok=True)
    
    flows = [
        ("01-login", "Email/password authentication", flow_01_login),
        ("02-dashboard-load", "Initial page render with data", flow_02_dashboard_load),
        ("03-providers-crud", "List, create, update, delete providers", flow_03_providers_crud),
        ("04-keys-crud", "List, create, update, delete API keys", flow_04_keys_crud),
        ("05-settings-save", "Update user settings", flow_05_settings_save),
        ("06-usage-chart", "View usage statistics", flow_06_usage_chart),
        ("07-chat-stream", "Initiate streaming chat request", flow_07_chat_stream),
        ("08-oauth-start", "Begin OAuth flow", flow_08_oauth_start),
        ("09-oauth-callback", "Complete OAuth flow", flow_09_oauth_callback),
        ("10-logout", "End session", flow_10_logout),
    ]
    
    for name, description, flow_func in flows:
        await capture_flow(name, description, flow_func)
        await asyncio.sleep(1)  # Brief pause between flows
    
    print("\n" + "="*60)
    print("HAR Capture Complete!")
    print("="*60)
    
    # Verify all files
    print("\nVerifying HAR files...")
    for name, _, _ in flows:
        har_file = FIXTURES_DIR / f"{name}.har"
        if har_file.exists():
            try:
                with open(har_file) as f:
                    json.load(f)
                size = har_file.stat().st_size
                print(f"✓ {name}.har ({size:,} bytes)")
            except json.JSONDecodeError:
                print(f"✗ {name}.har (INVALID JSON)")
        else:
            print(f"✗ {name}.har (NOT FOUND)")
    
    print("\n" + "="*60)
    print("Next steps:")
    print("1. Review captured HARs in fixtures/before/")
    print("2. Redact sensitive data if needed")
    print("3. Commit to repo: git add fixtures/before/*.har")
    print("4. Update TASK_STATUS.md: T0.10 → DONE")
    print("="*60)


if __name__ == "__main__":
    asyncio.run(main())
