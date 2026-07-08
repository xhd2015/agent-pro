// Playwright probe for agent-run web loop demo.
// Invoked via: playwright-debug run probe.js <baseURL> <outDir> <iteration> <prompt> [token] [runner]
//
// Prints a single REPORT_JSON=... line to stdout for the Go orchestrator.

const baseURL = (process.argv[3] || 'http://127.0.0.1:8192').replace(/\/$/, '');
const outDir = process.argv[4] || '/tmp/agent-run-web-loop';
const iteration = Number(process.argv[5] || '1');
const prompt = process.argv[6] || 'hello from web loop probe';
const apiToken = process.argv[7] || '';
const runner = process.argv[8] || 'grok-tty';

const fs = require('fs');
const path = require('path');

function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true });
}

async function shot(page, name) {
  const file = path.join(outDir, `iter-${iteration}-${name}.png`);
  await page.screenshot({ path: file, fullPage: true });
  return file;
}

async function countVisible(locator) {
  const n = await locator.count();
  let visible = 0;
  for (let i = 0; i < n; i++) {
    if (await locator.nth(i).isVisible().catch(() => false)) visible++;
  }
  return visible;
}

async function texts(locator) {
  const n = await locator.count();
  const out = [];
  for (let i = 0; i < n; i++) {
    const t = (await locator.nth(i).innerText().catch(() => '')).trim();
    if (t) out.push(t);
  }
  return out;
}

const report = {
  iteration,
  baseURL,
  prompt,
  screenshots: [],
  page: {},
  elements: {},
  network: { sessionDetailGets: 0, sseStreams: 0 },
  issues: [],
  ok: true,
};

page.on('request', (req) => {
  const url = req.url();
  if (!url.includes('/api/agent-run/sessions/')) return;
  if (url.includes('/events/stream')) {
    report.network.sseStreams++;
    return;
  }
  if (req.method() === 'GET' && /\/api\/agent-run\/sessions\/[^/]+\/[^/?]+(?:\?|$)/.test(url)) {
    report.network.sessionDetailGets++;
  }
});

// iPhone-class mobile viewport (matches web-layout doctests).
await page.setViewportSize({ width: 390, height: 844 });
report.viewport = { width: 390, height: 844 };

await page.addInitScript(({ token, runnerId }) => {
  if (token) localStorage.setItem('agent-run-token', token);
  if (runnerId) localStorage.setItem('agent-run-runner', runnerId);
}, { token: apiToken, runnerId: runner });
report.elements.runner = runner;

// --- Home ---
await page.goto(baseURL + '/', { waitUntil: 'domcontentloaded', timeout: 30000 });
report.screenshots.push(await shot(page, '01-home'));
report.page.url = page.url();
report.page.title = await page.title();

const authVisible = await page.locator('[data-testid="auth-page"]').isVisible().catch(() => false);
const emptyVisible = await page.locator('[data-testid="empty-state"]').isVisible().catch(() => false);
const composer = page.locator('[data-testid="composer"]');
const composerVisible = await composer.isVisible().catch(() => false);
report.elements.authPage = authVisible;
report.elements.emptyState = emptyVisible;
report.elements.composer = composerVisible;

if (authVisible) {
  report.ok = false;
  report.issues.push('auth-page visible: set --token auto or use open API (--token omitted on web)');
}

if (!composerVisible) {
  report.ok = false;
  report.issues.push('composer not visible on home');
}

// --- Send message via composer ---
if (composerVisible && !authVisible) {
  const input = page.locator('[data-testid="composer-input"]');
  await input.waitFor({ state: 'visible', timeout: 15000 });
  await input.fill(prompt);
  const sendBtn = page.locator('[data-testid="send-button"]');
  await sendBtn.click();

  // Wait for session route or chat surface
  try {
    await page.waitForURL(/\/sessions\//, { timeout: 30000 });
  } catch (e) {
    report.ok = false;
    report.issues.push('navigation to /sessions/... did not occur after send');
  }
}

report.screenshots.push(await shot(page, '02-after-send'));
report.page.urlAfterSend = page.url();
if (!report.page.urlAfterSend.includes('/sessions/')) {
  report.ok = false;
  report.issues.push('expected session URL to include runner from server default');
} else if (!report.page.urlAfterSend.includes(`/sessions/${runner}/`)) {
  report.issues.push(`session URL runner mismatch: wanted ${runner}, got ${report.page.urlAfterSend}`);
}

const chatActive = await page.locator('[data-testid="chat-active"]').isVisible().catch(() => false);
report.elements.chatActive = chatActive;

// Poll for user + assistant bubbles (grok mock ~2s; real grok longer)
const userLoc = page.locator('[data-testid="message-item-user"] .message-body');
const assistantLoc = page.locator('[data-testid="message-item-assistant"] .message-body');
const runningCard = page.locator('[data-testid="agent-running-card"]');

let userTexts = [];
let assistantTexts = [];
// grok-tty defers SSE ~1.5s after load; mock grok hook sleeps ~1s.
for (let i = 0; i < 80; i++) {
  userTexts = await texts(userLoc);
  assistantTexts = await texts(assistantLoc);
  if (userTexts.length >= 1 && assistantTexts.length >= 1) break;
  if (i === 79) break;
  await page.waitForTimeout(500);
}

report.screenshots.push(await shot(page, '03-after-wait'));
report.elements.userMessages = userTexts;
report.elements.assistantMessages = assistantTexts;
report.elements.userCount = userTexts.length;
report.elements.assistantCount = assistantTexts.length;
report.elements.runningCard = await runningCard.isVisible().catch(() => false);
report.elements.messageTimestamps = await countVisible(page.locator('[data-testid="message-timestamp"]'));

// Duplicate user text check
const userDupes = userTexts.filter((t, idx) => userTexts.indexOf(t) !== idx);
if (userDupes.length > 0) {
  report.ok = false;
  report.issues.push('duplicate user message texts: ' + [...new Set(userDupes)].join(' | '));
}

if (userTexts.length < 1) {
  report.ok = false;
  report.issues.push('no user message bubble after send');
}
if (assistantTexts.length < 1) {
  report.ok = false;
  report.issues.push('no assistant message bubble after send (timeout 30s)');
}

if (!userTexts.some((t) => t.includes(prompt) || prompt.includes(t.slice(0, 20)))) {
  if (userTexts.length > 0) {
    report.issues.push('user bubble text does not match prompt exactly (may be ok if truncated)');
  }
}

// Element inventory for agent inspection
report.elements.inventory = await page.evaluate(() => {
  const ids = [...document.querySelectorAll('[data-testid]')].map((el) => el.getAttribute('data-testid'));
  const counts = {};
  for (const id of ids) counts[id] = (counts[id] || 0) + 1;
  return counts;
});

// Mobile layout invariants (web-layout DOCTEST parity).
report.layout = await page.evaluate(() => {
  const docScroll = document.documentElement.scrollHeight - window.innerHeight;
  const composer = document.querySelector('[data-testid="composer"]');
  const messageList = document.querySelector('[data-testid="message-list"]');
  const composerBox = composer?.getBoundingClientRect();
  const vpH = window.innerHeight;
  const out = {
    documentScrollPx: docScroll,
    composerNearBottom: composerBox ? vpH - composerBox.bottom <= 48 : false,
    messageListScrollable: false,
    horizontalOverflow: document.documentElement.scrollWidth > window.innerWidth + 2,
  };
  if (messageList) {
    out.messageListScrollable = messageList.scrollHeight > messageList.clientHeight + 2;
  }
  return out;
});

if (report.layout.documentScrollPx > 4) {
  report.ok = false;
  report.issues.push(`document scrolls vertically (${report.layout.documentScrollPx}px); session page should be fixed chrome`);
}
if (!report.layout.composerNearBottom) {
  report.ok = false;
  report.issues.push('composer not pinned near viewport bottom on mobile');
}
if (report.layout.horizontalOverflow) {
  report.ok = false;
  report.issues.push('horizontal overflow on mobile viewport');
}
if (!report.elements.inventory?.['message-list']) {
  report.ok = false;
  report.issues.push('message-list missing on session page');
}

console.log('REPORT_JSON=' + JSON.stringify(report));