/**
 * Composer navigation probe for agent-run web home (verification step 7).
 * Usage:
 *   SCRATCH=/tmp/probe playwright-debug run script/debug/session-list-composer-nav-probe.js http://127.0.0.1:8192 /tmp/probe
 */
const fs = require('fs');
const path = require('path');

const baseURL = process.argv[3] || process.env.SESSION_LIST_URL || 'http://127.0.0.1:8192';
const SCRATCH = process.argv[4] || process.env.SCRATCH || path.dirname(__filename);
const logPath = path.join(SCRATCH, 'composer-nav.log');

await page.setViewportSize({ width: 390, height: 844 });
await page.goto(`${baseURL}/`, { waitUntil: 'networkidle', timeout: 30000 });
await page.waitForSelector('[data-testid="session-list"], [data-testid="empty-state"]', { timeout: 15000 });
await page.evaluate(() => {
  const list = document.querySelector('[data-testid="session-list"]');
  if (list) list.scrollTop = list.scrollHeight;
});
await page.waitForTimeout(200);

const composerPrompt = `Probe composer navigation ${Date.now()}`;
await page.fill('[data-testid="composer-input"]', composerPrompt);
await Promise.all([
  page.waitForURL(/\/sessions\/[^/]+\/[^/]+/, { timeout: 15000 }),
  page.click('[data-testid="send-button"]'),
]);

const navURL = page.url();
const line = `${navURL}\n`;
fs.writeFileSync(logPath, line);

if (!/\/sessions\/[^/]+\/[^/]+/.test(navURL)) {
  console.error('composer send did not navigate to session detail:', navURL);
  process.exit(1);
}

console.log('composer navigation probe passed');
console.log(line.trim());