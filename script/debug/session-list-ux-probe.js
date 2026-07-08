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
  flowMetrics.runningFilter = runningFilter;

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

await page.goto(`${baseURL}/`, { waitUntil: 'networkidle', timeout: 30000 });
await home.first().waitFor({ state: 'visible', timeout: 15000 });
await page.evaluate(() => {
  const list = document.querySelector('[data-testid="session-list"]');
  if (list) list.scrollTop = list.scrollHeight;
});
await page.waitForTimeout(200);
flowMetrics.final = await collectLayoutMetrics();

await page.screenshot({ path: screenshotPath, fullPage: false });

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