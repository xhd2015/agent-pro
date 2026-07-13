// Session list UX enhance probe (playwright-debug).
// Invoked via:
//   playwright-debug run probe.js <baseURL> <outDir> [token]
//
// Asserts three current symptoms (unfixed product):
//   1. Load more is outside the scrollable session list (sticky under list, not at list end)
//   2. Scrolling near bottom auto-loads the next page (no explicit button click)
//   3. Enter session → back loses list scroll (and often loaded pages)
//
// Healthy (after fix):
//   1. Load more is inside the scroll container (at end of list content)
//   2. Scroll alone does not increase item count; button click does
//   3. After session → back, scrollTop is preserved (and loaded pages kept)
//
// Prints one REPORT_JSON=... line. process.exit(0) always from playwright.

const baseURL = (process.argv[3] || 'http://127.0.0.1:8192').replace(/\/$/, '');
const outDir = process.argv[4] || '/tmp/session-list-ux-enhance';
const apiToken = process.argv[5] || '';

const fs = require('fs');
const path = require('path');

fs.mkdirSync(outDir, { recursive: true });

const MOBILE = { width: 390, height: 844 };
const issues = [];
const screenshots = [];
const snapshots = {};

async function shot(name) {
  const file = path.join(outDir, `${name}.png`);
  await page.screenshot({ path: file, fullPage: false });
  screenshots.push(file);
  return file;
}

async function seedTokenIfNeeded() {
  if (!apiToken) return;
  await page.addInitScript((tok) => {
    try {
      localStorage.setItem('agent-run-token', tok);
    } catch (_) {}
  }, apiToken);
}

async function readListState() {
  return page.evaluate(() => {
    const list = document.querySelector('[data-testid="session-list"]');
    const btn = document.querySelector('[data-testid="session-load-more"]');
    const items = Array.from(document.querySelectorAll('[data-testid="session-item"]'));
    const firstId =
      items[0]?.getAttribute('href') ||
      items[0]?.querySelector('[data-testid="session-preview"]')?.textContent ||
      null;
    const lastId =
      items[items.length - 1]?.getAttribute('href') ||
      items[items.length - 1]?.querySelector('[data-testid="session-preview"]')?.textContent ||
      null;

    let btnRect = null;
    let btnVisibleInViewport = false;
    if (btn) {
      const r = btn.getBoundingClientRect();
      btnRect = {
        top: r.top,
        bottom: r.bottom,
        height: r.height,
        width: r.width,
      };
      btnVisibleInViewport =
        r.height > 0 && r.bottom > 0 && r.top < window.innerHeight && r.width > 0;
    }

    const btnInsideList = !!(btn && list && list.contains(btn));
    // "At end of list content" also allows a single scroll parent that wraps
    // both list items and the load-more button.
    let loadMoreInScrollContainer = btnInsideList;
    let scrollContainer = list;
    if (btn && list && !btnInsideList) {
      // Walk up from list; if load-more shares a scrollable ancestor with list
      // and is after the list in document order, treat as "in scroll container".
      let el = list.parentElement;
      while (el && el !== document.body) {
        const style = getComputedStyle(el);
        const oy = style.overflowY;
        const scrollable =
          (oy === 'auto' || oy === 'scroll' || oy === 'overlay') &&
          el.scrollHeight > el.clientHeight + 2;
        if (scrollable && el.contains(btn) && el.contains(list)) {
          loadMoreInScrollContainer = true;
          scrollContainer = el;
          break;
        }
        el = el.parentElement;
      }
    }

    return {
      itemCount: items.length,
      hasLoadMore: !!btn,
      btnInsideList,
      loadMoreInScrollContainer,
      btnParentClass: btn?.parentElement?.className || null,
      btnVisibleInViewport,
      btnRect,
      listScrollTop: list?.scrollTop ?? null,
      listScrollHeight: list?.scrollHeight ?? null,
      listClientHeight: list?.clientHeight ?? null,
      listOverflow: list ? list.scrollHeight > list.clientHeight + 2 : false,
      scrollContainerIsList: scrollContainer === list,
      firstId,
      lastId,
      path: location.pathname,
    };
  });
}

function classify(symptom) {
  const reasons = [];
  if (symptom.loadMoreNotAtListEnd) {
    reasons.push('load more button is not at the end of the session list scroll content');
  }
  if (symptom.autoLoadOnScroll) {
    reasons.push(
      `scroll near bottom auto-loaded pages (${symptom.itemCountBeforeScroll} -> ${symptom.itemCountAfterScroll})`,
    );
  }
  if (symptom.scrollLostOnBack) {
    reasons.push(
      `session enter/back lost scroll (saved=${symptom.scrollTopBeforeNav} after=${symptom.scrollTopAfterBack})`,
    );
  }
  if (symptom.scrollSnapBack) {
    reasons.push(
      `multi-step scroll snapped back (steps=${JSON.stringify(symptom.scrollSteps)} final=${symptom.scrollAfterSettle} expected≈${symptom.scrollStepC})`,
    );
  }
  if (symptom.wastefulMetaPoll) {
    reasons.push(
      `home idle poll refetches runners/status (runnersΔ=${symptom.apiHitsAfterIdle?.runnersDelta} statusΔ=${symptom.apiHitsAfterIdle?.statusDelta}; sessionsΔ=${symptom.apiHitsAfterIdle?.sessionsDelta})`,
    );
  }
  return reasons;
}

/** Simulate scroll gestures with wheel intent + settle gaps (repro mid-list snap-back). */
async function scrollListToFraction(frac) {
  return page.evaluate((f) => {
    const list = document.querySelector('[data-testid="session-list"]');
    if (!list) return 0;
    const max = Math.max(0, list.scrollHeight - list.clientHeight);
    const target = Math.floor(max * f);
    // Dispatch wheel first so markUserScrollIntent runs (same as real trackpad start).
    list.dispatchEvent(
      new WheelEvent('wheel', { deltaY: target > list.scrollTop ? 40 : -40, bubbles: true }),
    );
    list.scrollTop = target;
    list.dispatchEvent(new Event('scroll', { bubbles: true }));
    return list.scrollTop;
  }, frac);
}

await seedTokenIfNeeded();
await page.setViewportSize(MOBILE);

// Track home bootstrap + periodic API noise (user report: runners/status polled every 3s).
const apiHits = { sessions: 0, runners: 0, status: 0, other: 0, urls: [] };
page.on('request', (req) => {
  if (req.method() !== 'GET') return;
  const u = req.url();
  if (!u.includes('/api/agent-run/')) return;
  apiHits.urls.push(u.replace(baseURL, ''));
  if (u.includes('/api/agent-run/sessions')) apiHits.sessions += 1;
  else if (u.includes('/api/agent-run/runners')) apiHits.runners += 1;
  else if (u.includes('/api/agent-run/status')) apiHits.status += 1;
  else apiHits.other += 1;
});

await page.goto(`${baseURL}/`, { waitUntil: 'networkidle', timeout: 45000 });
await page.waitForSelector('[data-testid="session-list"]', { timeout: 20000 });
// Allow first page paint + any initial fetch.
await page.waitForTimeout(400);

// Sit idle ~7s: old code hit runners+status every 3s (~2–3 extras). Healthy: one bootstrap each.
const apiHitsAfterBootstrap = {
  sessions: apiHits.sessions,
  runners: apiHits.runners,
  status: apiHits.status,
};
await page.waitForTimeout(7000);
const apiHitsAfterIdle = {
  sessions: apiHits.sessions,
  runners: apiHits.runners,
  status: apiHits.status,
  sessionsDelta: apiHits.sessions - apiHitsAfterBootstrap.sessions,
  runnersDelta: apiHits.runners - apiHitsAfterBootstrap.runners,
  statusDelta: apiHits.status - apiHitsAfterBootstrap.status,
};
snapshots.apiHitsAfterBootstrap = apiHitsAfterBootstrap;
snapshots.apiHitsAfterIdle = apiHitsAfterIdle;
// Unfixed: runners/status re-fetched on the 3s poll during idle.
const wastefulMetaPoll =
  apiHitsAfterIdle.runnersDelta >= 1 || apiHitsAfterIdle.statusDelta >= 1;

const initial = await readListState();
snapshots.initial = initial;
await shot('01-home-top');

if (!initial.listOverflow && initial.itemCount < 30) {
  issues.push(
    `session list does not overflow (itemCount=${initial.itemCount}); need more seeded sessions`,
  );
}
if (!initial.hasLoadMore && initial.itemCount >= 30) {
  // has_more may still be true with button present only when hasMore; if total==page
  // this is infra. Prefer continue with best-effort.
  issues.push('load more button missing despite a full first page');
}

// --- Symptom 1: load more only at end of list (inside scroll content) ---
// Unfixed: button is sibling under .session-list-region, always in viewport at scrollTop=0.
const loadMoreNotAtListEnd =
  initial.hasLoadMore &&
  (!initial.loadMoreInScrollContainer ||
    (initial.btnVisibleInViewport &&
      (initial.listScrollTop ?? 0) < 8 &&
      initial.listOverflow &&
      !initial.btnInsideList));

// --- Symptom 2: no auto-load on scroll ---
// Scroll list to bottom without clicking load more; wait for possible auto-fetch.
const itemCountBeforeScroll = initial.itemCount;
const sessionFetches = [];
page.on('request', (req) => {
  const u = req.url();
  if (u.includes('/api/agent-run/sessions') && req.method() === 'GET') {
    sessionFetches.push({ url: u, t: Date.now() });
  }
});

await page.evaluate(() => {
  const list = document.querySelector('[data-testid="session-list"]');
  if (!list) return;
  list.scrollTop = list.scrollHeight;
});
await page.waitForTimeout(900);
const afterScroll = await readListState();
snapshots.afterScroll = afterScroll;
await shot('02-after-scroll-near-bottom');

const itemCountAfterScroll = afterScroll.itemCount;
const autoLoadOnScroll = itemCountAfterScroll > itemCountBeforeScroll + 2;
// Also detect offset>0 fetch during the wait (polls may hit offset=0).
const offsetFetches = sessionFetches.filter((f) => /[?&]offset=([1-9]\d*)/.test(f.url));
const autoLoadViaFetch = offsetFetches.length > 0 && autoLoadOnScroll;

// --- Symptom 4: multi-step scroll must not snap back after settle / poll ---
// User report: scroll A → hang → B → hang → C stay → later jumps back to B.
const scrollSteps = [];
const stepA = await scrollListToFraction(0.25);
scrollSteps.push({ name: 'A', top: stepA });
await page.waitForTimeout(400);
const stepASettle = await page.evaluate(
  () => document.querySelector('[data-testid="session-list"]')?.scrollTop ?? 0,
);
scrollSteps.push({ name: 'A_settle', top: stepASettle });
const stepB = await scrollListToFraction(0.45);
scrollSteps.push({ name: 'B', top: stepB });
await page.waitForTimeout(400);
const stepBSettle = await page.evaluate(
  () => document.querySelector('[data-testid="session-list"]')?.scrollTop ?? 0,
);
scrollSteps.push({ name: 'B_settle', top: stepBSettle });
const stepC = await scrollListToFraction(0.7);
scrollSteps.push({ name: 'C', top: stepC });
// Wait past poll interval (3s) + margin so content refresh would re-pin if buggy.
await page.waitForTimeout(3500);
const scrollAfterSettle = await page.evaluate(
  () => document.querySelector('[data-testid="session-list"]')?.scrollTop ?? 0,
);
scrollSteps.push({ name: 'C_after_poll', top: scrollAfterSettle });
snapshots.scrollSteps = scrollSteps;
await shot('02b-multi-step-scroll-settle');

const scrollSnapBack =
  stepC >= 80 &&
  (Math.abs(scrollAfterSettle - stepC) > 80 ||
    (stepBSettle >= 80 &&
      Math.abs(scrollAfterSettle - stepBSettle) < 40 &&
      Math.abs(scrollAfterSettle - stepC) > 80));

// --- Symptom 3: preserve scroll when entering session and back ---
// Reset list to a mid scroll position on current loaded list.
await page.evaluate(() => {
  const list = document.querySelector('[data-testid="session-list"]');
  if (!list) return;
  // Detach from top: not following "latest" (top-anchored list).
  const mid = Math.max(120, Math.floor((list.scrollHeight - list.clientHeight) * 0.45));
  list.scrollTop = mid;
  list.dispatchEvent(new Event('scroll', { bubbles: true }));
});
await page.waitForTimeout(300);
const beforeNav = await readListState();
snapshots.beforeNav = beforeNav;
const scrollTopBeforeNav = beforeNav.listScrollTop ?? 0;
const itemCountBeforeNav = beforeNav.itemCount;
await shot('03-scrolled-before-nav');

if (scrollTopBeforeNav < 40) {
  issues.push(`could not establish mid scroll before nav (scrollTop=${scrollTopBeforeNav})`);
}

// Click a mid-list session if possible (not first). force:true avoids jump-to-latest overlay.
// Do NOT scrollIntoViewIfNeeded — that would change list scrollTop before navigate.
const itemCount = await page.locator('[data-testid="session-item"]').count();
const clickIndex = itemCount > 3 ? Math.min(4, itemCount - 1) : 0;
const targetItem = page.locator('[data-testid="session-item"]').nth(clickIndex);
// Ensure mid scroll is still applied, fire scroll + mousedown for persist handlers.
await page.evaluate((idx) => {
  const list = document.querySelector('[data-testid="session-list"]');
  if (!list) return;
  const mid = Math.max(120, Math.floor((list.scrollHeight - list.clientHeight) * 0.45));
  list.scrollTop = mid;
  list.dispatchEvent(new Event('scroll', { bubbles: true }));
  const items = list.querySelectorAll('[data-testid="session-item"]');
  const item = items[idx];
  if (item) {
    item.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, cancelable: true }));
  }
}, clickIndex);
await page.waitForTimeout(150);
const scrollTopRightBeforeClick = await page.evaluate(() => {
  return document.querySelector('[data-testid="session-list"]')?.scrollTop ?? 0;
});
snapshots.scrollTopRightBeforeClick = scrollTopRightBeforeClick;
await targetItem.click({ force: true });
await page.waitForURL(/\/sessions\//, { timeout: 15000 });
await page.waitForTimeout(300);
snapshots.onSession = { path: new URL(page.url()).pathname };
await shot('04-session-detail');

// Back to home via product back control.
const back = page.locator('.back-link').first();
await back.waitFor({ state: 'visible', timeout: 10000 });
await back.click();
await page.waitForURL((url) => {
  try {
    return new URL(url).pathname === '/';
  } catch {
    return false;
  }
}, { timeout: 15000 });
await page.waitForSelector('[data-testid="session-list"]', { timeout: 15000 });
// Allow restore layout + rAF timeouts (0/50/200ms) to re-apply scroll.
await page.waitForTimeout(500);
const afterBack = await readListState();
snapshots.afterBack = afterBack;
await shot('05-after-back');

const scrollTopAfterBack = afterBack.listScrollTop ?? 0;
const itemCountAfterBack = afterBack.itemCount;
// Lost if scroll reset near 0 while we had a meaningful mid position.
const scrollLostOnBack =
  scrollTopBeforeNav >= 80 && Math.abs(scrollTopAfterBack - scrollTopBeforeNav) > 60;

const symptom = {
  loadMoreNotAtListEnd,
  autoLoadOnScroll: autoLoadOnScroll || autoLoadViaFetch,
  scrollLostOnBack,
  scrollSnapBack,
  wastefulMetaPoll,
  apiHitsAfterBootstrap,
  apiHitsAfterIdle,
  scrollSteps,
  scrollStepC: stepC,
  scrollAfterSettle,
  itemCountBeforeScroll,
  itemCountAfterScroll,
  offsetFetchCount: offsetFetches.length,
  scrollTopBeforeNav,
  scrollTopAfterBack,
  itemCountBeforeNav,
  itemCountAfterBack,
  btnInsideListInitial: initial.btnInsideList,
  loadMoreInScrollContainerInitial: initial.loadMoreInScrollContainer,
  btnVisibleAtTop: initial.btnVisibleInViewport && (initial.listScrollTop ?? 0) < 8,
};

const reasons = classify(symptom);
const symptomPresent =
  symptom.loadMoreNotAtListEnd ||
  symptom.autoLoadOnScroll ||
  symptom.scrollLostOnBack ||
  symptom.scrollSnapBack ||
  symptom.wastefulMetaPoll;

// Extra check: load more click still works (explicit trigger).
let loadMoreClick = null;
{
  // Mid scroll first so hasMore button still exists after prior scroll tests.
  await page.evaluate(() => {
    const list = document.querySelector('[data-testid="session-list"]');
    if (list) list.scrollTop = list.scrollHeight;
    let el = list?.parentElement;
    while (el && el !== document.body) {
      const style = getComputedStyle(el);
      if (
        (style.overflowY === 'auto' || style.overflowY === 'scroll') &&
        el.scrollHeight > el.clientHeight
      ) {
        el.scrollTop = el.scrollHeight;
      }
      el = el.parentElement;
    }
  });
  await page.waitForTimeout(250);
  const beforeClick = await readListState();
  if (beforeClick.hasLoadMore) {
    const countBefore = beforeClick.itemCount;
    await page.locator('[data-testid="session-load-more"]').click({ force: true, timeout: 10000 });
    await page.waitForTimeout(900);
    const afterClick = await readListState();
    loadMoreClick = {
      before: countBefore,
      after: afterClick.itemCount,
      increased: afterClick.itemCount > countBefore,
    };
    snapshots.afterLoadMoreClick = afterClick;
    await shot('06-after-load-more-click');
    if (!loadMoreClick.increased) {
      issues.push(
        `load more click did not increase items (${countBefore} -> ${afterClick.itemCount})`,
      );
    }
  } else if (itemCountBeforeScroll >= 30 && itemCountAfterScroll <= itemCountBeforeScroll) {
    // has_more expected for seed; if already fully loaded via bug, note it.
    issues.push('load more button missing after scroll (cannot verify explicit click load)');
  }
}

// Healthy requires UX criteria fixed + load-more click works when present.
const okHealthy =
  !symptom.loadMoreNotAtListEnd &&
  !symptom.autoLoadOnScroll &&
  !symptom.scrollLostOnBack &&
  !symptom.scrollSnapBack &&
  !symptom.wastefulMetaPoll &&
  apiHitsAfterBootstrap.runners <= 2 &&
  apiHitsAfterBootstrap.status <= 2 &&
  initial.hasLoadMore &&
  afterBack.itemCount >= itemCountBeforeNav - 2 &&
  (loadMoreClick == null || loadMoreClick.increased) &&
  issues.length === 0;

const report = {
  baseURL,
  symptomPresent,
  okHealthy,
  issues,
  symptom,
  reasons,
  snapshots,
  screenshots,
  loadMoreClick,
  sessionFetches: sessionFetches.slice(-20),
  page: { viewport: MOBILE },
};

fs.writeFileSync(path.join(outDir, 'probe-report.json'), JSON.stringify(report, null, 2));
console.log('REPORT_JSON=' + JSON.stringify(report));
// Always exit 0; orchestrator interprets symptomPresent / okHealthy.
process.exit(0);
