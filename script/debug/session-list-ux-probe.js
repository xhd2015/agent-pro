/**
 * Playwright probe for agent-run web home session-list UX.
 * Usage:
 *   SCRATCH=/tmp/probe playwright-debug run script/debug/session-list-ux-probe.js http://127.0.0.1:8192 /tmp/probe 1
 */
const fs = require('fs');
const path = require('path');

const baseURL = process.argv[3] || process.env.SESSION_LIST_URL || 'http://127.0.0.1:8192';
const SCRATCH = process.argv[4] || process.env.SCRATCH || path.dirname(__filename);
const runNum = process.argv[5] || process.env.PROBE_RUN || '1';
const screenshotPath = path.join(SCRATCH, `session-list-ux-${runNum}.png`);
const reportPath = path.join(SCRATCH, 'session-list-report.json');
const runReportPath = path.join(SCRATCH, `session-list-report-${runNum}.json`);

const pageErrors = [];
page.on('pageerror', (err) => pageErrors.push(String(err)));

await page.setViewportSize({ width: 390, height: 844 });
await page.goto(`${baseURL}/`, { waitUntil: 'networkidle', timeout: 30000 });

const home = page.locator('[data-testid="home-active"], [data-testid="empty-state"]');
await home.first().waitFor({ state: 'visible', timeout: 15000 });

const issues = [];

const metrics = await page.evaluate(() => {
  const doc = document.documentElement;
  const body = document.body;
  const docScroll = Math.max(doc.scrollHeight - doc.clientHeight, body.scrollHeight - body.clientHeight);
  const docOverflowX = Math.max(doc.scrollWidth - doc.clientWidth, body.scrollWidth - body.clientWidth);

  const sessionList = document.querySelector('[data-testid="session-list"]');
  const composer = document.querySelector('[data-testid="composer"]');
  const runnerPicker = document.querySelector('[data-testid="runner-picker"]');
  const composerRect = composer?.getBoundingClientRect();
  const composerBottomGap = composerRect ? window.innerHeight - composerRect.bottom : null;
  const runnerVisible = runnerPicker
    ? runnerPicker.getBoundingClientRect().height > 0 && getComputedStyle(runnerPicker).visibility !== 'hidden'
    : false;

  const items = Array.from(document.querySelectorAll('[data-testid="session-item"]'));
  const rows = items.map((el) => {
    const preview = (el.querySelector('[data-testid="session-preview"]')?.textContent || '').trim();
    const status = (el.querySelector('[data-testid="session-status"]')?.textContent || '').trim();
    const recency = (el.querySelector('[data-testid="session-recency"]')?.textContent || '').trim();
    const rect = el.getBoundingClientRect();
    const statusEl = el.querySelector('[data-testid="session-status"]');
    const statusStyle = statusEl ? getComputedStyle(statusEl) : null;
    return {
      preview,
      status,
      recency,
      height: rect.height,
      statusBackground: statusStyle?.backgroundColor || null,
      statusColor: statusStyle?.color || null,
    };
  });

  return {
    docScroll,
    docOverflowX,
    itemCount: items.length,
    rows,
    sessionListVisible: Boolean(sessionList),
    composerBottomGap,
    runnerVisible,
    viewportHeight: window.innerHeight,
    emptyStateVisible: Boolean(document.querySelector('[data-testid="empty-state"]')),
    homeActiveVisible: Boolean(document.querySelector('[data-testid="home-active"]')),
  };
});

if (pageErrors.length > 0) {
  issues.push(`page errors: ${pageErrors.join('; ')}`);
}
if (!metrics.homeActiveVisible && !metrics.emptyStateVisible) {
  issues.push('neither home-active nor empty-state visible');
}
if (metrics.itemCount > 0) {
  if (!metrics.sessionListVisible) {
    issues.push('session-list not visible with populated items');
  }
  for (const [i, row] of metrics.rows.entries()) {
    if (!row.preview) issues.push(`session-item[${i}] missing preview/label text`);
    if (!row.status) issues.push(`session-item[${i}] missing status pill`);
    if (!row.recency) issues.push(`session-item[${i}] missing recency text`);
    if (row.height < 44) issues.push(`session-item[${i}] tap target too short (${row.height}px)`);
  }
}
if (metrics.docOverflowX > 2) {
  issues.push(`horizontal overflow ${metrics.docOverflowX}px`);
}
if (metrics.docScroll > 4) {
  issues.push(`document scroll ${metrics.docScroll}px exceeds 4px`);
}
if (!metrics.runnerVisible) {
  issues.push('runner-picker not visible on mobile viewport');
}
if (metrics.composerBottomGap == null || metrics.composerBottomGap > 48) {
  issues.push(`composer not pinned near bottom (gap=${metrics.composerBottomGap}px)`);
}

await page.screenshot({ path: screenshotPath, fullPage: false });

const report = {
  run: runNum,
  url: `${baseURL}/`,
  issues,
  metrics,
  pass: issues.length === 0,
  screenshot: screenshotPath,
};
const runJson = JSON.stringify(report, null, 2);
fs.writeFileSync(runReportPath, runJson);

let combined = { url: report.url, runs: [report], pass: report.pass };
try {
  if (fs.existsSync(reportPath)) {
    const prev = JSON.parse(fs.readFileSync(reportPath, 'utf8'));
    if (Array.isArray(prev.runs)) {
      const withoutDup = prev.runs.filter((r) => String(r.run) !== String(runNum));
      combined.runs = [...withoutDup, report];
      combined.pass = combined.runs.every((r) => r.pass);
    }
  }
} catch {
  combined = { url: report.url, runs: [report], pass: report.pass };
}
fs.writeFileSync(reportPath, JSON.stringify(combined, null, 2));

if (issues.length > 0) {
  console.error('UX issues:', issues);
  process.exit(1);
}

console.log('session-list UX probe passed');
console.log(runJson);