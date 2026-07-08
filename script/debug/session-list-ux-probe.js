/**
 * Playwright probe for agent-run web home session-list UX.
 * Usage:
 *   SCRATCH=/tmp/probe playwright-debug run script/debug/session-list-ux-probe.js http://127.0.0.1:8192 /tmp/probe 1
 */
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const baseURL = process.argv[3] || process.env.SESSION_LIST_URL || 'http://127.0.0.1:8192';
const SCRATCH = process.argv[4] || process.env.SCRATCH || path.dirname(__filename);
const runNum = process.argv[5] || process.env.PROBE_RUN || '1';
const screenshotPath = path.join(SCRATCH, `session-list-ux-${runNum}.png`);
const reportPath = path.join(SCRATCH, 'session-list-report.json');
const runReportPath = path.join(SCRATCH, `session-list-report-${runNum}.json`);
const pageErrors = [];
page.on('pageerror', (err) => pageErrors.push(String(err)));

function distanceFromBottom(el) {
  return el.scrollHeight - el.scrollTop - el.clientHeight;
}

async function collectLayoutMetrics() {
  return page.evaluate(() => {
    const doc = document.documentElement;
    const body = document.body;
    const docScroll = Math.max(doc.scrollHeight - doc.clientHeight, body.scrollHeight - body.clientHeight);
    const docOverflowX = Math.max(doc.scrollWidth - doc.clientWidth, body.scrollWidth - body.clientWidth);

    const sessionList = document.querySelector('[data-testid="session-list"]');
    const composer = document.querySelector('[data-testid="composer"]');
    const runnerPicker = document.querySelector('[data-testid="runner-picker"]');
    const topBar = document.querySelector('.top-bar-home');
    const composerRect = composer?.getBoundingClientRect();
    const composerBottomGap = composerRect ? window.innerHeight - composerRect.bottom : null;
    const runnerVisible = runnerPicker
      ? runnerPicker.getBoundingClientRect().height > 0 && getComputedStyle(runnerPicker).visibility !== 'hidden'
      : false;

    const items = Array.from(document.querySelectorAll('[data-testid="session-item"]'));
    const rows = items.map((el) => {
      const preview = (el.querySelector('[data-testid="session-preview"]')?.textContent || '').trim();
      const statusEl = el.querySelector('[data-testid="session-status"]');
      const status = (statusEl?.getAttribute('data-status') || statusEl?.textContent || '').trim();
      const recency = (el.querySelector('[data-testid="session-recency"]')?.textContent || '').trim();
      const runner = (el.querySelector('[data-testid="session-runner"]')?.textContent || '').trim();
      const workspace = (el.querySelector('[data-testid="session-workspace"]')?.textContent || '').trim();
      const rect = el.getBoundingClientRect();
      const statusStyle = statusEl ? getComputedStyle(statusEl) : null;
      const itemStyle = getComputedStyle(el);
      return {
        preview,
        status,
        recency,
        runner,
        workspace,
        height: rect.height,
        statusBackground: statusStyle?.backgroundColor || null,
        statusColor: statusStyle?.color || null,
        borderLeftColor: itemStyle.borderLeftColor || null,
        itemClass: el.className,
      };
    });

    return {
      docScroll,
      docOverflowX,
      itemCount: items.length,
      rows,
      sessionListVisible: Boolean(sessionList),
      sessionListOverflow: sessionList ? sessionList.scrollHeight > sessionList.clientHeight + 2 : false,
      sessionListScrollTop: sessionList?.scrollTop ?? null,
      composerBottomGap,
      runnerVisible,
      topBarY: topBar?.getBoundingClientRect().top ?? null,
      composerY: composer?.getBoundingClientRect().top ?? null,
      viewportHeight: window.innerHeight,
      emptyStateVisible: Boolean(document.querySelector('[data-testid="empty-state"]')),
      homeLoadingVisible: Boolean(document.querySelector('[data-testid="home-loading"]')),
      homeActiveVisible: Boolean(document.querySelector('[data-testid="home-active"]')),
      sessionListHeaderVisible: Boolean(document.querySelector('[data-testid="session-list-header"]')),
      sessionFilterChipsVisible: Boolean(document.querySelector('[data-testid="session-filter-chips"]')),
      jumpToLatestVisible: Boolean(document.querySelector('[data-testid="jump-to-latest"]')),
    };
  });
}

function validateBaseMetrics(metrics, issues) {
  if (!metrics.homeActiveVisible && !metrics.emptyStateVisible && !metrics.homeLoadingVisible) {
    issues.push('neither home-active, empty-state, nor home-loading visible');
  }
  if (metrics.homeLoadingVisible && metrics.itemCount > 0) {
    issues.push('home-loading visible while session rows are rendered');
  }
  if (metrics.itemCount > 0) {
    if (!metrics.sessionListVisible) {
      issues.push('session-list not visible with populated items');
    }
    if (!metrics.sessionListHeaderVisible) {
      issues.push('session-list-header not visible');
    }
    if (!metrics.sessionFilterChipsVisible) {
      issues.push('session-filter-chips not visible');
    }
    for (const [i, row] of metrics.rows.entries()) {
      if (!row.preview) issues.push(`session-item[${i}] missing preview/label text`);
      if (!row.status) issues.push(`session-item[${i}] missing status pill`);
      if (!row.recency) issues.push(`session-item[${i}] missing recency text`);
      if (!row.runner) issues.push(`session-item[${i}] missing runner label`);
      if (!row.workspace) issues.push(`session-item[${i}] missing workspace label`);
      if (row.height < 44) issues.push(`session-item[${i}] tap target too short (${row.height}px)`);
    }
    const runningRows = metrics.rows.filter((row) => row.status === 'running');
    const finishedRows = metrics.rows.filter((row) => row.status === 'finished' || row.status === 'idle');
    if (runningRows.length === 0) {
      issues.push('expected at least one running session row for resume-running UX');
    }
    if (finishedRows.length === 0) {
      issues.push('expected at least one finished/idle session row for contrast');
    }
    if (runningRows.length > 0 && finishedRows.length > 0) {
      const running = runningRows[0];
      const finished = finishedRows[0];
      const distinctStatus =
        running.statusBackground !== finished.statusBackground ||
        running.statusColor !== finished.statusColor;
      const distinctBorder = running.borderLeftColor !== finished.borderLeftColor;
      if (!distinctStatus && !distinctBorder) {
        issues.push('running and finished rows lack visual distinction');
      }
      if (!running.itemClass.includes('session-item--running')) {
        issues.push('running row missing session-item--running class');
      }
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
}

await page.setViewportSize({ width: 390, height: 844 });
await page.goto(`${baseURL}/`, { waitUntil: 'domcontentloaded', timeout: 30000 });

const home = page.locator('[data-testid="home-active"], [data-testid="empty-state"]');
await home.first().waitFor({ state: 'visible', timeout: 15000 });

const issues = [];
const flowMetrics = {};

const chromeOnLoad = await page.evaluate(() => ({
  topBarVisible: Boolean(document.querySelector('.top-bar-home')),
  composerVisible: Boolean(document.querySelector('[data-testid="composer"]')),
  mainPanelChildCount: document.querySelector('[data-testid="home-active"]')?.childElementCount ?? 0,
}));
if (!chromeOnLoad.topBarVisible) issues.push('top bar not visible on initial load');
if (!chromeOnLoad.composerVisible) issues.push('composer not visible on initial load');
flowMetrics.chromeOnLoad = chromeOnLoad;

await page.waitForLoadState('networkidle', { timeout: 30000 }).catch(() => {});

const ensureRunningSession = await page.evaluate(async (url) => {
  const res = await fetch(`${url}/api/agent-run/sessions`);
  if (!res.ok) return { ok: false, reason: `sessions status ${res.status}` };
  const data = await res.json();
  const statuses = (data.sessions || []).map((s) => s.status);
  if (statuses.includes('running')) return { ok: true, reseeded: false, statuses };
  const created = await fetch(`${url}/api/agent-run/sessions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      runner: 'opencode',
      prompt: 'Probe re-seed running session for UX probe',
    }),
  });
  if (!created.ok) return { ok: false, reason: `create status ${created.status}` };
  return { ok: true, reseeded: true };
}, baseURL);
if (!ensureRunningSession.ok) {
  issues.push(`ensure running session failed: ${ensureRunningSession.reason}`);
} else if (ensureRunningSession.reseeded) {
  await page.waitForTimeout(2000);
  await page.reload({ waitUntil: 'networkidle', timeout: 30000 });
}
flowMetrics.ensureRunningSession = ensureRunningSession;

let metrics = await collectLayoutMetrics();
validateBaseMetrics(metrics, issues);
flowMetrics.initial = metrics;

if (metrics.itemCount > 0 && metrics.sessionFilterChipsVisible) {
  const allCount = metrics.itemCount;
  await page.click('[data-testid="session-filter-running"]');
  await page.waitForTimeout(150);
  const runningFilter = await page.evaluate(() => {
    const items = Array.from(document.querySelectorAll('[data-testid="session-item"]'));
    const statuses = items.map(
      (el) => el.querySelector('[data-testid="session-status"]')?.getAttribute('data-status') || '',
    );
    return { itemCount: items.length, statuses, active: document.querySelector('[data-testid="session-filter-running"]')?.className || '' };
  });
  if (!runningFilter.active.includes('session-filter-chip--active')) {
    issues.push('running filter chip not active after click');
  }
  if (runningFilter.itemCount === 0) {
    issues.push('running filter shows zero rows despite seeded running session');
  }
  if (runningFilter.statuses.some((s) => s !== 'running')) {
    issues.push(`running filter includes non-running statuses: ${runningFilter.statuses.join(',')}`);
  }
  if (runningFilter.itemCount >= allCount) {
    issues.push(`running filter did not reduce visible rows (${runningFilter.itemCount} >= ${allCount})`);
  }
  const runningCountText = (
    await page.locator('[data-testid="session-list-count"]').innerText()
  ).trim();
  if (!new RegExp(`^${runningFilter.itemCount} sessions? running$`).test(runningCountText)) {
    issues.push(`running filter count label mismatch: "${runningCountText}" vs ${runningFilter.itemCount}`);
  }
  flowMetrics.runningFilter = { ...runningFilter, countText: runningCountText };

  await page.click('[data-testid="session-filter-done"]');
  await page.waitForTimeout(150);
  const doneFilter = await page.evaluate(() => {
    const chipCount = Number(
      document.querySelector('[data-testid="session-filter-done"] .session-filter-chip-count')?.textContent || '0',
    );
    const items = Array.from(document.querySelectorAll('[data-testid="session-item"]'));
    const statuses = items.map(
      (el) => el.querySelector('[data-testid="session-status"]')?.getAttribute('data-status') || '',
    );
    return { chipCount, itemCount: items.length, statuses };
  });
  if (doneFilter.chipCount !== doneFilter.itemCount) {
    issues.push(`done chip count ${doneFilter.chipCount} != visible rows ${doneFilter.itemCount}`);
  }
  if (doneFilter.statuses.some((s) => s !== 'finished' && s !== 'idle')) {
    issues.push(`done filter includes non-done statuses: ${doneFilter.statuses.join(',')}`);
  }
  flowMetrics.doneFilter = doneFilter;

  await page.click('[data-testid="session-filter-all"]');
  await page.waitForTimeout(150);
  const allFilter = await collectLayoutMetrics();
  if (allFilter.itemCount !== allCount) {
    issues.push(`all filter itemCount mismatch (${allFilter.itemCount} vs ${allCount})`);
  }
  flowMetrics.allFilterRestore = { itemCount: allFilter.itemCount };
}

if (metrics.sessionListVisible && metrics.sessionListOverflow) {
  const beforeScroll = await collectLayoutMetrics();
  await page.evaluate(() => {
    const list = document.querySelector('[data-testid="session-list"]');
    if (!list) return;
    list.scrollTop = Math.max(0, list.scrollHeight - list.clientHeight - 250);
  });
  await page.waitForTimeout(200);
  const afterScroll = await collectLayoutMetrics();
  if (Math.abs((afterScroll.topBarY ?? 0) - (beforeScroll.topBarY ?? 0)) > 2) {
    issues.push('top bar moved while session-list scrolled');
  }
  if (Math.abs((afterScroll.composerY ?? 0) - (beforeScroll.composerY ?? 0)) > 2) {
    issues.push('composer moved while session-list scrolled');
  }
  flowMetrics.sessionsOnlyScroll = { beforeScroll, afterScroll };

  const detached = await page.evaluate(() => {
    const list = document.querySelector('[data-testid="session-list"]');
    if (!list) return { distance: 0, scrollTop: 0 };
    return { distance: list.scrollHeight - list.scrollTop - list.clientHeight, scrollTop: list.scrollTop };
  });
  if (detached.distance <= 80) {
    issues.push(`session-list not detached after scroll-up (distance=${detached.distance})`);
  }

  const scrollTopBeforePoll = detached.scrollTop;
  await page.waitForTimeout(3500);
  const afterPoll = await page.evaluate(() => {
    const list = document.querySelector('[data-testid="session-list"]');
    if (!list) return { scrollTop: 0, distance: 0, jumpVisible: false };
    return {
      scrollTop: list.scrollTop,
      distance: list.scrollHeight - list.scrollTop - list.clientHeight,
      jumpVisible: Boolean(document.querySelector('[data-testid="jump-to-latest"]')),
    };
  });
  if (Math.abs(afterPoll.scrollTop - scrollTopBeforePoll) > 2) {
    issues.push(`detached scrollTop changed after poll (${scrollTopBeforePoll} -> ${afterPoll.scrollTop})`);
  }
  flowMetrics.detachPoll = { scrollTopBeforePoll, afterPoll };

  if (afterPoll.jumpVisible) {
    await page.click('[data-testid="session-filter-running"]');
    await page.waitForTimeout(200);
    const detachFilterChange = await page.evaluate(() => ({
      jumpVisible: Boolean(document.querySelector('[data-testid="jump-to-latest"]')),
      filterEmpty: Boolean(document.querySelector('[data-testid="session-filter-empty"]')),
      countText: (document.querySelector('[data-testid="session-list-count"]')?.textContent || '').trim(),
      itemCount: document.querySelectorAll('[data-testid="session-item"]').length,
      listExists: Boolean(document.querySelector('[data-testid="session-list"]')),
    }));
    if (detachFilterChange.jumpVisible) {
      issues.push('jump-to-latest still visible after filter change while detached');
    }
    if (detachFilterChange.filterEmpty) {
      issues.push('filter-empty shown after switching to running while detached');
    }
    if (!detachFilterChange.countText.includes('running')) {
      issues.push(`detached filter change count label missing running: ${detachFilterChange.countText}`);
    }
    flowMetrics.detachFilterChange = detachFilterChange;

    await page.click('[data-testid="session-filter-all"]');
    await page.waitForTimeout(200);
    await page.evaluate(() => {
      const list = document.querySelector('[data-testid="session-list"]');
      if (list) list.scrollTop = Math.max(0, list.scrollHeight - list.clientHeight - 250);
    });
    await page.waitForTimeout(3500);
    const jumpAgain = await page.evaluate(() => ({
      jumpVisible: Boolean(document.querySelector('[data-testid="jump-to-latest"]')),
    }));
    if (!jumpAgain.jumpVisible) {
      issues.push('jump-to-latest not visible after re-detach on all filter');
    }

    const jumpShotPath = path.join(SCRATCH, `session-list-jump-to-latest-${runNum}.png`);
    await page.locator('[data-testid="jump-to-latest"]').screenshot({ path: jumpShotPath });
    if (!fs.statSync(jumpShotPath).size) {
      issues.push('screenshot jump-to-latest is empty');
    }
    flowMetrics.jumpScreenshot = jumpShotPath;
    await page.click('[data-testid="jump-to-latest"]');
    await page.waitForTimeout(200);
    const afterJump = await page.evaluate(() => {
      const list = document.querySelector('[data-testid="session-list"]');
      if (!list) return { distance: 999, jumpVisible: true };
      return {
        distance: list.scrollHeight - list.scrollTop - list.clientHeight,
        jumpVisible: Boolean(document.querySelector('[data-testid="jump-to-latest"]')),
      };
    });
    if (afterJump.distance > 80) {
      issues.push(`jump-to-latest did not reach bottom (distance=${afterJump.distance})`);
    }
    if (afterJump.jumpVisible) {
      issues.push('jump-to-latest chip still visible after tap');
    }
    flowMetrics.jumpToLatest = afterJump;
  } else {
    issues.push('jump-to-latest chip not visible while detached after poll');
  }
} else if (metrics.itemCount > 0) {
  issues.push('session-list does not overflow; cannot verify scroll-only and jump-to-latest flows');
}

const surfaceScreenshots = {};

async function captureSurface(name, selector) {
  const outPath = path.join(SCRATCH, `session-list-${name}-${runNum}.png`);
  const locator = page.locator(selector).first();
  await locator.waitFor({ state: 'visible', timeout: 10000 });
  await locator.screenshot({ path: outPath });
  const stat = fs.statSync(outPath);
  if (!stat.size) {
    issues.push(`screenshot ${name} is empty`);
  }
  surfaceScreenshots[name] = outPath;
  return outPath;
}

async function captureComposerSurface() {
  const outPath = path.join(SCRATCH, `session-list-composer-${runNum}.png`);
  await page.locator('[data-testid="composer"]').screenshot({ path: outPath });
  const stat = fs.statSync(outPath);
  if (!stat.size) issues.push('screenshot composer is empty');
  surfaceScreenshots.composer = outPath;
  return outPath;
}

if (metrics.itemCount > 0) {
  await captureSurface('top-bar', '.top-bar-home');
  await captureSurface('header-filters', '[data-testid="session-list-header"]');
  if (await page.locator('[data-testid="quick-resume-strip"]').isVisible()) {
    await captureSurface('quick-resume', '[data-testid="quick-resume-strip"]');
  }
  await page.evaluate(() => {
    const list = document.querySelector('[data-testid="session-list"]');
    if (list) list.scrollTop = list.scrollHeight;
  });
  await page.waitForTimeout(200);
  await captureSurface('session-rows', '[data-testid="session-list"]');
  await captureComposerSurface();

  const activeBadge = await page.evaluate(() => {
    const badge = document.querySelector('[data-testid="session-running-badge"]');
    return {
      tag: badge?.tagName?.toLowerCase() ?? null,
      activeFilter: document.querySelector('[data-testid="session-filter-running"]')?.className ?? '',
    };
  });
  if (activeBadge.tag !== 'button') {
    issues.push('session-running-badge is not a button');
  }
  await page.click('[data-testid="session-running-badge"]');
  await page.waitForTimeout(150);
  const afterBadge = await page.evaluate(() => ({
    activeFilter: document.querySelector('[data-testid="session-filter-running"]')?.className ?? '',
    itemCount: document.querySelectorAll('[data-testid="session-item"]').length,
    statuses: Array.from(document.querySelectorAll('[data-testid="session-status"]')).map(
      (el) => el.getAttribute('data-status') || '',
    ),
  }));
  if (!afterBadge.activeFilter.includes('session-filter-chip--active')) {
    issues.push('active badge click did not activate running filter');
  }
  if (afterBadge.statuses.some((s) => s !== 'running')) {
    issues.push(`active badge filter includes non-running: ${afterBadge.statuses.join(',')}`);
  }
  flowMetrics.activeBadgeFilter = afterBadge;
  await page.click('[data-testid="session-filter-all"]');
  await page.waitForTimeout(150);
}

if (metrics.sessionListVisible && metrics.sessionListOverflow && !flowMetrics.jumpScreenshot) {
  issues.push('jump-to-latest screenshot was not captured during detach flow');
}

const agentRunHome = process.env.AGENT_RUN_HOME || '';
if (agentRunHome) {
  try {
    execSync(`bash "${path.join(__dirname, 'finish-all-running-sessions.sh')}"`, {
      env: { ...process.env, AGENT_RUN_HOME: agentRunHome },
      stdio: 'pipe',
    });
    flowMetrics.finishAllRunning = true;
  } catch (err) {
    issues.push(`finish-all-running-sessions failed: ${err.stderr?.toString() || err.message}`);
  }
} else {
  issues.push('AGENT_RUN_HOME unset; cannot exercise real filter-empty with poll');
}

await page.goto(`${baseURL}/`, { waitUntil: 'domcontentloaded', timeout: 30000 });
await home.first().waitFor({ state: 'visible', timeout: 15000 });
await page.waitForTimeout(3500);
await page.click('[data-testid="session-filter-running"]');
await page.waitForTimeout(200);
const realFilterEmpty = await page.evaluate(() => ({
  filterEmpty: Boolean(document.querySelector('[data-testid="session-filter-empty"]')),
  jumpVisible: Boolean(document.querySelector('[data-testid="jump-to-latest"]')),
  countText: (document.querySelector('[data-testid="session-list-count"]')?.textContent || '').trim(),
  runningChipCount: Number(
    document.querySelector('[data-testid="session-filter-running"] .session-filter-chip-count')?.textContent || '0',
  ),
}));
flowMetrics.realFilterEmpty = realFilterEmpty;
if (!realFilterEmpty.filterEmpty) {
  issues.push('session-filter-empty not visible after finishing all running sessions');
}
if (realFilterEmpty.jumpVisible) {
  issues.push('jump-to-latest visible on real filter-empty state');
}
if (realFilterEmpty.runningChipCount !== 0) {
  issues.push(`running chip count not zero after finish-all-running: ${realFilterEmpty.runningChipCount}`);
}
if (realFilterEmpty.filterEmpty) {
  await captureSurface('filter-empty', '[data-testid="session-filter-empty"]');
}

await page.route('**/api/agent-run/sessions', async (route) => {
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({ sessions: [] }),
  });
});
await page.goto(`${baseURL}/`, { waitUntil: 'domcontentloaded', timeout: 30000 });
await page.waitForSelector('[data-testid="empty-state"]', { timeout: 15000 });
await page.waitForTimeout(300);
await captureSurface('empty-state', '[data-testid="empty-state"]');
await captureComposerSurface();
await page.unroute('**/api/agent-run/sessions').catch(() => {});

await page.goto(`${baseURL}/`, { waitUntil: 'networkidle', timeout: 30000 });
await home.first().waitFor({ state: 'visible', timeout: 15000 });
await page.evaluate(() => {
  const list = document.querySelector('[data-testid="session-list"]');
  if (list) list.scrollTop = list.scrollHeight;
});
await page.waitForTimeout(200);
flowMetrics.final = await collectLayoutMetrics();

await page.screenshot({ path: screenshotPath, fullPage: false });
if (flowMetrics.jumpScreenshot) {
  surfaceScreenshots['jump-to-latest'] = flowMetrics.jumpScreenshot;
}
flowMetrics.surfaceScreenshots = surfaceScreenshots;

if (pageErrors.length > 0) {
  issues.push(`page errors: ${pageErrors.join('; ')}`);
}

const report = {
  run: runNum,
  url: `${baseURL}/`,
  issues,
  metrics: flowMetrics.initial ?? metrics,
  flowMetrics,
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