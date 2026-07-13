// Playwright repro for message-card reorder on grok-tty follow-up.
// Invoked via:
//   playwright-debug run probe.js <baseURL> <outDir> [token]
//
// User flow:
//   1. open home  2. choose grok-tty  3. send "run ls"
//   4. wait finished  5. send "what did I say"
// Symptom (live, pre-refresh): first user card "run ls" jumps so it sits
// immediately above "what did I say" with no assistant card between them
// (or assistant for turn 1 appears above both user cards). Refresh restores order.
//
// Prints one REPORT_JSON=... line. process.exit(0) always from playwright
// (orchestrator interprets report.symptomPresent / report.okHealthy).

const baseURL = (process.argv[3] || 'http://127.0.0.1:8192').replace(/\/$/, '');
const outDir = process.argv[4] || '/tmp/message-card-reorder-followup';
const apiToken = process.argv[5] || '';

const PROMPT1 = 'run ls';
const PROMPT2 = 'what did I say';

const fs = require('fs');
const path = require('path');

fs.mkdirSync(outDir, { recursive: true });

async function shot(name) {
  const file = path.join(outDir, `${name}.png`);
  await page.screenshot({ path: file, fullPage: true });
  return file;
}

async function readTimeline() {
  return page.evaluate(() => {
    const nodes = Array.from(
      document.querySelectorAll(
        '[data-testid="message-item-user"], [data-testid="message-item-assistant"]',
      ),
    );
    return nodes.map((el) => {
      const testid = el.getAttribute('data-testid') || '';
      const role = testid.includes('user') ? 'user' : 'assistant';
      const text = (el.querySelector('.message-body')?.textContent || '').trim();
      return { role, text };
    });
  });
}

function userTexts(timeline) {
  return timeline.filter((e) => e.role === 'user').map((e) => e.text);
}

function findUserIndex(timeline, needle) {
  return timeline.findIndex(
    (e) => e.role === 'user' && (e.text.includes(needle) || needle.includes(e.text)),
  );
}

/**
 * Symptom (product bug): after follow-up, first user card jumps below assistants
 * (strip-all-users merge) or user order inverts.
 *
 * Not a bug: adjacent users when turn-1 never produced an assistant (mock bind race),
 * as long as first user still precedes assistants and second user.
 */
function detectAdjacentUserReorder(timeline, opts = {}) {
  const { hadAssistantAfterFirst = false } = opts;
  const i1 = findUserIndex(timeline, PROMPT1);
  const i2 = findUserIndex(timeline, PROMPT2);
  if (i1 < 0 || i2 < 0) {
    return { hit: false, reason: 'missing user prompts in timeline' };
  }
  if (i1 > i2) {
    return {
      hit: true,
      reason: `user order inverted: "${PROMPT1}" at ${i1} after "${PROMPT2}" at ${i2}`,
      i1,
      i2,
    };
  }
  // Primary signature of strip-all-users + re-append: assistant before first user.
  const firstAssistant = timeline.findIndex((e) => e.role === 'assistant' && e.text.length > 0);
  if (firstAssistant >= 0 && firstAssistant < i1) {
    return {
      hit: true,
      reason: `assistant at ${firstAssistant} appears before first user "${PROMPT1}" at ${i1}`,
      i1,
      i2,
      firstAssistant,
    };
  }
  // If turn 1 had an assistant, follow-up must keep it between the two user cards.
  const between = timeline.slice(i1 + 1, i2);
  const hasAssistantBetween = between.some((e) => e.role === 'assistant' && e.text.length > 0);
  if (hadAssistantAfterFirst && !hasAssistantBetween) {
    return {
      hit: true,
      reason: `"${PROMPT1}" lost assistant between it and "${PROMPT2}" (live reorder)`,
      i1,
      i2,
      between,
    };
  }
  return { hit: false, reason: 'healthy user order', i1, i2, firstAssistant };
}

async function statusText() {
  return page.evaluate(() => {
    const pill = document.querySelector('[class*="status"], .status-pill, [data-testid="session-status"]');
    // session header status pill has classes like status-finished / status-running
    const candidates = Array.from(document.querySelectorAll('.session-header-row span, .top-bar span'));
    for (const el of candidates) {
      const t = (el.textContent || '').trim().toLowerCase();
      if (t === 'running' || t === 'finished' || t === 'idle' || t === 'error') return t;
    }
    const running = document.querySelector('[data-testid="agent-running-card"]');
    if (running) return 'running';
    return '';
  });
}

async function waitForUserPrompt(text, timeoutMs) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const tl = await readTimeline();
    if (findUserIndex(tl, text) >= 0) return tl;
    await page.waitForTimeout(250);
  }
  return readTimeline();
}

async function waitForAssistantAfterUser(userText, timeoutMs) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const tl = await readTimeline();
    const ui = findUserIndex(tl, userText);
    if (ui >= 0) {
      const after = tl.slice(ui + 1);
      if (after.some((e) => e.role === 'assistant' && e.text.length > 0 && e.text !== '…')) {
        return tl;
      }
    }
    await page.waitForTimeout(300);
  }
  return readTimeline();
}

async function waitNotRunning(timeoutMs) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    const st = await statusText();
    const runningCard = await page.locator('[data-testid="agent-running-card"]').isVisible().catch(() => false);
    const loading = await page
      .locator('[data-testid="message-item-assistant-loading"]')
      .isVisible()
      .catch(() => false);
    if (!runningCard && !loading && st !== 'running') {
      return st || 'idle';
    }
    await page.waitForTimeout(300);
  }
  return statusText();
}

const report = {
  baseURL,
  prompts: { first: PROMPT1, second: PROMPT2 },
  screenshots: [],
  snapshots: {},
  symptomPresent: false,
  symptom: null,
  okHealthy: false,
  issues: [],
  page: {},
};

await page.setViewportSize({ width: 390, height: 844 });

await page.addInitScript((token) => {
  if (token) localStorage.setItem('agent-run-token', token);
  localStorage.setItem('agent-run-runner', 'grok-tty');
}, apiToken);

// --- Home: choose grok-tty ---
await page.goto(baseURL + '/', { waitUntil: 'domcontentloaded', timeout: 30000 });
report.screenshots.push(await shot('01-home'));

const authVisible = await page.locator('[data-testid="auth-page"]').isVisible().catch(() => false);
if (authVisible) {
  report.issues.push('auth-page visible; need open API or token');
  console.log('REPORT_JSON=' + JSON.stringify(report));
  process.exit(0);
}

const runnerSelect = page.locator('[data-testid="runner-select"]');
await runnerSelect.waitFor({ state: 'visible', timeout: 15000 });
await runnerSelect.selectOption('grok-tty').catch(async () => {
  // fallback: set value + change event
  await page.evaluate(() => {
    const el = document.querySelector('[data-testid="runner-select"]');
    if (!el) return;
    el.value = 'grok-tty';
    el.dispatchEvent(new Event('change', { bubbles: true }));
  });
});
const selected = await runnerSelect.inputValue().catch(() => '');
report.page.runnerSelected = selected;
if (selected !== 'grok-tty') {
  report.issues.push(`runner-select value=${selected}, expected grok-tty`);
}

const composer = page.locator('[data-testid="composer-input"]');
await composer.waitFor({ state: 'visible', timeout: 15000 });

// --- Send first prompt ---
await composer.fill(PROMPT1);
await page.locator('[data-testid="send-button"]').click();
try {
  await page.waitForURL(/\/sessions\//, { timeout: 30000 });
} catch {
  report.issues.push('no navigation to /sessions after first send');
}
report.page.urlAfterFirst = page.url();
report.screenshots.push(await shot('02-after-first-send'));

await page.locator('[data-testid="chat-active"]').waitFor({ state: 'visible', timeout: 20000 }).catch(() => {});

// First turn may take ~20s if grok session bind races; still show user bubble after timeout.
let tl = await waitForUserPrompt(PROMPT1, 45000);
// Prefer assistant after first user, but do not hard-fail if only bind errors appear.
tl = await waitForAssistantAfterUser(PROMPT1, 30000);
const afterFirstStatus = await waitNotRunning(60000);
const i1AfterFirst = findUserIndex(tl, PROMPT1);
const hadAssistantAfterFirst =
  i1AfterFirst >= 0 &&
  tl.slice(i1AfterFirst + 1).some((e) => e.role === 'assistant' && e.text.length > 0);
report.snapshots.afterFirstTurn = {
  status: afterFirstStatus,
  timeline: tl,
  users: userTexts(tl),
  hadAssistantAfterFirst,
};
report.screenshots.push(await shot('03-after-first-finished'));

if (findUserIndex(tl, PROMPT1) < 0) {
  report.issues.push('first user prompt never appeared');
  console.log('REPORT_JSON=' + JSON.stringify(report));
  process.exit(0);
}

// --- Send second prompt (follow-up) ---
const composer2 = page.locator('[data-testid="composer-input"]');
await composer2.waitFor({ state: 'visible', timeout: 15000 });
// Wait until send is enabled (not mid-send)
for (let i = 0; i < 40; i++) {
  const disabled = await page.locator('[data-testid="send-button"]').isDisabled().catch(() => true);
  if (!disabled) break;
  await page.waitForTimeout(250);
}
await composer2.fill(PROMPT2);
await page.locator('[data-testid="send-button"]').click();

// Poll for symptom during follow-up (bug often appears right after optimistic merge / refresh).
const pollDeadline = Date.now() + 60000;
let sawSecondUser = false;
let firstHit = null;
let everHealthyLive = false;
const pollTrace = [];
while (Date.now() < pollDeadline) {
  tl = await readTimeline();
  const users = userTexts(tl);
  if (findUserIndex(tl, PROMPT2) >= 0) sawSecondUser = true;
  const det = detectAdjacentUserReorder(tl, { hadAssistantAfterFirst });
  pollTrace.push({
    t: Date.now(),
    users,
    roles: tl.map((e) => e.role + ':' + (e.text || '').slice(0, 40)),
    hit: det.hit,
    reason: det.reason,
  });
  if (det.hit && sawSecondUser) {
    firstHit = { ...det, timeline: tl, users };
    break;
  }
  if (sawSecondUser && !det.hit) {
    everHealthyLive = true;
  }
  const st = await statusText();
  const runningCard = await page.locator('[data-testid="agent-running-card"]').isVisible().catch(() => false);
  // Once second user is visible, not running, and healthy — can stop early for verify.
  if (sawSecondUser && !runningCard && st !== 'running' && !det.hit && Date.now() > pollDeadline - 45000) {
    // require a short settle after second user
    await page.waitForTimeout(800);
    tl = await readTimeline();
    const det2 = detectAdjacentUserReorder(tl, { hadAssistantAfterFirst });
    if (!det2.hit) break;
  }
  await page.waitForTimeout(250);
}

report.snapshots.liveAfterSecond = {
  sawSecondUser,
  timeline: tl,
  users: userTexts(tl),
  symptom: firstHit,
  everHealthyLive,
  hadAssistantAfterFirst,
  pollSamples: pollTrace.length,
  lastPolls: pollTrace.slice(-8),
};
report.screenshots.push(await shot('04-live-after-second'));

if (firstHit) {
  report.symptomPresent = true;
  report.symptom = firstHit;
} else if (sawSecondUser) {
  const det = detectAdjacentUserReorder(tl, { hadAssistantAfterFirst });
  if (det.hit) {
    report.symptomPresent = true;
    report.symptom = { ...det, timeline: tl };
  } else {
    // Live path healthy — expected after fix
    report.symptomPresent = false;
  }
} else {
  report.issues.push('second user prompt never appeared in timeline');
}

// --- Reload (expected healthy order) ---
await page.reload({ waitUntil: 'domcontentloaded', timeout: 30000 });
await page.locator('[data-testid="chat-active"]').waitFor({ state: 'visible', timeout: 20000 }).catch(() => {});
// allow hydration
await page.waitForTimeout(2000);
for (let i = 0; i < 20; i++) {
  tl = await readTimeline();
  if (findUserIndex(tl, PROMPT1) >= 0 && findUserIndex(tl, PROMPT2) >= 0) break;
  await page.waitForTimeout(500);
}
const afterReloadDet = detectAdjacentUserReorder(tl, { hadAssistantAfterFirst: false });
report.snapshots.afterReload = {
  timeline: tl,
  users: userTexts(tl),
  reorderHit: afterReloadDet.hit,
  reason: afterReloadDet.reason,
};
report.screenshots.push(await shot('05-after-reload'));

// Healthy: both users present, chronological, no assistant-before-first-user (live + reload).
const liveFinal = detectAdjacentUserReorder(
  report.snapshots.liveAfterSecond.timeline || [],
  { hadAssistantAfterFirst },
);
report.okHealthy =
  findUserIndex(tl, PROMPT1) >= 0 &&
  findUserIndex(tl, PROMPT2) >= 0 &&
  !afterReloadDet.hit &&
  !report.symptomPresent &&
  !liveFinal.hit;

fs.writeFileSync(path.join(outDir, 'report.json'), JSON.stringify(report, null, 2));
console.log('REPORT_JSON=' + JSON.stringify(report));
