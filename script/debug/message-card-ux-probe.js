/**
 * Playwright probe for agent-run web message-card UX on a live session URL.
 * Usage:
 *   SCRATCH=/tmp/probe AGENT_RUN_TOKEN=<token> PROBE_RUN=1 \
 *     playwright-debug run script/debug/message-card-ux-probe.js
 */
const fs = require('fs');
const path = require('path');

const SCRATCH = process.env.SCRATCH || process.cwd();
const runNum = process.env.PROBE_RUN || process.argv.slice(3).find((a) => /^\d+$/.test(a)) || '1';
const screenshotPath = path.join(SCRATCH, `message-card-ux-${runNum}.png`);
const reportPath = path.join(SCRATCH, 'message-card-report.json');

const url =
  process.env.MESSAGE_CARD_URL ||
  'http://127.0.0.1:8192/sessions/grok-tty/web_a1e886dbcebb3e2b';
const token = process.env.AGENT_RUN_TOKEN || '';

await page.setViewportSize({ width: 390, height: 844 });

await page.goto('http://127.0.0.1:8192/', { waitUntil: 'domcontentloaded', timeout: 30000 });
if (token) {
  await page.evaluate((t) => localStorage.setItem('agent-run-token', t), token);
}

await page.goto(url, { waitUntil: 'networkidle', timeout: 30000 });

const chatActive = page.locator('[data-testid="chat-active"]');
await chatActive.waitFor({ state: 'visible', timeout: 15000 });
await page.waitForTimeout(1500);

const issues = [];

const metrics = await page.evaluate(() => {
  const doc = document.documentElement;
  const body = document.body;
  const docScroll = Math.max(doc.scrollHeight - doc.clientHeight, body.scrollHeight - body.clientHeight);
  const docOverflowX = Math.max(doc.scrollWidth - doc.clientWidth, body.scrollWidth - body.clientWidth);

  const userCards = Array.from(document.querySelectorAll('[data-testid="message-item-user"]'));
  const assistantCards = Array.from(document.querySelectorAll('[data-testid="message-item-assistant"]'));

  const cardStyle = (el) => {
    const s = getComputedStyle(el);
    const r = el.getBoundingClientRect();
    return {
      background: s.backgroundColor,
      textAlign: s.textAlign,
      alignSelf: s.alignSelf,
      left: r.left,
      right: window.innerWidth - r.right,
      width: r.width,
    };
  };

  const bodyText = (el) => (el.querySelector('.message-body')?.textContent || '').trim();
  const composer = document.querySelector('[data-testid="composer"]');
  const composerRect = composer?.getBoundingClientRect();
  const composerBottomGap = composerRect ? window.innerHeight - composerRect.bottom : null;

  return {
    docScroll,
    docOverflowX,
    userCount: userCards.length,
    assistantCount: assistantCards.length,
    progressCount: document.querySelectorAll('[data-testid="progress-card"]').length,
    userBodies: userCards.map(bodyText).filter(Boolean),
    assistantBodies: assistantCards.map(bodyText).filter(Boolean),
    userStyles: userCards.map(cardStyle),
    assistantStyles: assistantCards.map(cardStyle),
    composerBottomGap,
    roleLabels: Array.from(document.querySelectorAll('.message-role')).map((el) => el.textContent?.trim()),
  };
});

if (metrics.userCount === 0 && metrics.assistantCount === 0) {
  issues.push('no user or assistant message cards found');
}
if (metrics.userBodies.length === 0 && metrics.assistantBodies.length === 0) {
  issues.push('no non-empty .message-body text');
}
if (metrics.userCount > 0 && metrics.assistantCount > 0) {
  const u = metrics.userStyles[0];
  const a = metrics.assistantStyles[0];
  const distinctBg = u.background !== a.background;
  const distinctAlign = u.left > a.left + 20 || a.left > u.left + 20;
  if (!distinctBg && !distinctAlign) {
    issues.push('user and assistant cards lack visual distinction');
  }
}
if (metrics.roleLabels.length < 2) {
  issues.push(`expected readable role labels, got ${metrics.roleLabels.length}`);
}
if (metrics.docOverflowX > 2) {
  issues.push(`horizontal overflow ${metrics.docOverflowX}px`);
}
if (metrics.docScroll > 4) {
  issues.push(`document scroll ${metrics.docScroll}px exceeds 4px`);
}
if (metrics.composerBottomGap == null || metrics.composerBottomGap > 48) {
  issues.push(`composer not pinned near bottom (gap=${metrics.composerBottomGap}px)`);
}

const overflowCheck = await page.evaluate(() => {
  const bad = [];
  for (const el of document.querySelectorAll('[data-testid="message-item-user"], [data-testid="message-item-assistant"]')) {
    const r = el.getBoundingClientRect();
    if (r.right > window.innerWidth + 2) bad.push(`card overflows right by ${r.right - window.innerWidth}px`);
    if (r.left < -2) bad.push(`card overflows left by ${Math.abs(r.left)}px`);
  }
  return bad;
});
issues.push(...overflowCheck);

if (metrics.progressCount > 0) {
  const progressDistinct = await page.evaluate(() => {
    const progress = document.querySelector('[data-testid="progress-card"]');
    const bubble = document.querySelector('[data-testid="message-item-assistant"], [data-testid="message-item-user"]');
    if (!progress || !bubble) return true;
    const ps = getComputedStyle(progress);
    const bs = getComputedStyle(bubble);
    return ps.backgroundColor !== bs.backgroundColor || ps.borderRadius !== bs.borderRadius;
  });
  if (!progressDistinct) {
    issues.push('progress cards not visually distinct from message bubbles');
  }
}

await page.screenshot({ path: screenshotPath, fullPage: false });

const report = { run: runNum, url, issues, metrics, pass: issues.length === 0, screenshot: screenshotPath };
fs.writeFileSync(reportPath, JSON.stringify(report, null, 2));

if (issues.length > 0) {
  console.error('UX issues:', issues);
  process.exit(1);
}

console.log('message-card UX probe passed');
console.log(JSON.stringify(report, null, 2));