// Mobile terminal touch-scroll probe (playwright-debug).
// Invoked via:
//   playwright-debug run probe.js <baseURL> <outDir> [token] [sessionPath]
//
// Mobile viewport + synthetic touch pan on [data-testid=terminal-surface].
// Symptom (unfixed): touch pan does not reveal older LINE_xxx scrollback.
// Healthy: touch pan decreases min visible LINE index (history scrolls).
//
// Prints one REPORT_JSON=... line. process.exit(0) always from playwright.

const baseURL = (process.argv[3] || 'http://127.0.0.1:8192').replace(/\/$/, '');
const outDir = process.argv[4] || '/tmp/mobile-terminal-touch-scroll';
const apiToken = process.argv[5] || '';
const sessionPath = process.argv[6] || '/sessions/web_term_touch_scroll';

const fs = require('fs');
const path = require('path');

fs.mkdirSync(outDir, { recursive: true });

const MOBILE = { width: 390, height: 844 };

async function shot(name) {
  const file = path.join(outDir, `${name}.png`);
  await page.screenshot({ path: file, fullPage: true });
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

/** Parse LINE_NNN markers from terminal visible text. */
function parseLineMarkers(text) {
  const re = /LINE_(\d{3})/g;
  const nums = [];
  let m;
  while ((m = re.exec(text || '')) !== null) {
    nums.push(parseInt(m[1], 10));
  }
  if (nums.length === 0) {
    return { count: 0, min: null, max: null, sample: [] };
  }
  return {
    count: nums.length,
    min: Math.min(...nums),
    max: Math.max(...nums),
    sample: nums.slice(0, 8),
  };
}

async function readTerminalSnapshot() {
  return page.evaluate(() => {
    const root = document.querySelector('[data-testid="terminal-surface"]');
    if (!root) {
      return { ok: false, reason: 'no terminal-surface' };
    }
    const rows = root.querySelector('.xterm-rows');
    const text = (rows && rows.textContent) || root.textContent || '';
    // Best-effort scrollable DOM metric (xterm v5 viewport or v6 scrollable).
    let scrollTop = null;
    let scrollHeight = null;
    let clientHeight = null;
    const all = root.querySelectorAll('*');
    let bestRange = 0;
    for (const el of all) {
      const sh = el.scrollHeight || 0;
      const ch = el.clientHeight || 0;
      if (sh > ch + 10 && sh - ch > bestRange) {
        bestRange = sh - ch;
        scrollTop = el.scrollTop;
        scrollHeight = sh;
        clientHeight = ch;
      }
    }
    return {
      ok: true,
      text: text.slice(0, 4000),
      textLen: text.length,
      hasXterm: !!root.querySelector('.xterm'),
      scrollTop,
      scrollHeight,
      clientHeight,
      scrollRange: bestRange,
    };
  });
}

/**
 * Synthetic single-finger vertical pan on the terminal surface.
 * Finger moves down (positive clientY delta) → expect older scrollback.
 */
async function touchPanVertical(opts = {}) {
  const {
    // finger travels this many px down the screen
    distancePx = 280,
    steps = 12,
  } = opts;

  return page.evaluate(
    ({ distancePx, steps }) => {
      const root = document.querySelector('[data-testid="terminal-surface"]');
      if (!root) throw new Error('no terminal-surface for touch pan');
      const rect = root.getBoundingClientRect();
      const x = rect.left + rect.width / 2;
      const startY = rect.top + rect.height * 0.35;
      const endY = startY + distancePx;

      const fire = (type, y, touchesList) => {
        const touch = new Touch({
          identifier: 1,
          target: root,
          clientX: x,
          clientY: y,
          pageX: x,
          pageY: y,
          screenX: x,
          screenY: y,
          radiusX: 2,
          radiusY: 2,
          rotationAngle: 0,
          force: 1,
        });
        const init = {
          bubbles: true,
          cancelable: true,
          composed: true,
          touches: touchesList(touch),
          targetTouches: touchesList(touch),
          changedTouches: [touch],
        };
        root.dispatchEvent(new TouchEvent(type, init));
      };

      fire('touchstart', startY, (t) => [t]);
      for (let i = 1; i <= steps; i++) {
        const y = startY + ((endY - startY) * i) / steps;
        fire('touchmove', y, (t) => [t]);
      }
      fire('touchend', endY, () => []);

      return { x, startY, endY, distancePx, steps };
    },
    { distancePx, steps },
  );
}

/** Desktop-style wheel (for comparison; may still work under mobile viewport). */
async function wheelScroll(deltaY = -600) {
  const surface = page.locator('[data-testid="terminal-surface"]');
  const box = await surface.boundingBox();
  if (!box) throw new Error('no bounding box for wheel');
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
  await page.mouse.wheel(0, deltaY);
}

const screenshots = [];
const report = {
  baseURL,
  sessionPath,
  mobile: MOBILE,
  symptomPresent: false,
  okHealthy: false,
  issues: [],
  symptom: {},
  snapshots: {},
  screenshots: [],
  page: {},
  gesture: {},
};

try {
  await page.setViewportSize(MOBILE);
  await seedTokenIfNeeded();

  const url = baseURL + sessionPath;
  await page.goto(url, { waitUntil: 'networkidle', timeout: 45000 });
  await page.waitForSelector('[data-testid="chat-active"], [data-testid="message-list"]', {
    timeout: 20000,
  });

  const terminalButton = page.getByRole('button', { name: /terminal/i });
  await terminalButton.waitFor({ state: 'visible', timeout: 20000 });
  await terminalButton.click();

  const surface = page.locator('[data-testid="terminal-surface"]');
  await surface.waitFor({ state: 'visible', timeout: 20000 });
  await page.locator('[data-testid="terminal-surface"] .xterm').waitFor({
    state: 'visible',
    timeout: 20000,
  });

  // Wait for scrollback markers from fake PTY.
  await page.waitForFunction(
    () => {
      const el = document.querySelector('[data-testid="terminal-surface"] .xterm-rows');
      const t = (el && el.textContent) || '';
      return /LINE_\d{3}/.test(t);
    },
    null,
    { timeout: 20000 },
  );

  // Brief settle after fit/resize writes.
  await page.waitForTimeout(400);
  screenshots.push(await shot('01-terminal-open'));

  const beforeSnap = await readTerminalSnapshot();
  const beforeLines = parseLineMarkers(beforeSnap.text || '');
  report.snapshots.before = { ...beforeSnap, lines: beforeLines };

  if (beforeLines.count === 0) {
    report.issues.push('no LINE_xxx markers visible after terminal open');
  }

  // Touch pan: finger down → expect older (lower) LINE indices if scroll works.
  const gesture = await touchPanVertical({ distancePx: 320, steps: 16 });
  report.gesture.touch1 = gesture;
  await page.waitForTimeout(200);
  // Second pan for more travel when residual/cell mapping is coarse.
  report.gesture.touch2 = await touchPanVertical({ distancePx: 320, steps: 16 });
  await page.waitForTimeout(250);

  screenshots.push(await shot('02-after-touch-pan'));

  const afterTouchSnap = await readTerminalSnapshot();
  const afterTouchLines = parseLineMarkers(afterTouchSnap.text || '');
  report.snapshots.afterTouch = { ...afterTouchSnap, lines: afterTouchLines };

  // Optional wheel baseline (not required for symptom, useful diagnostics).
  let wheelMoved = false;
  try {
    await wheelScroll(-800);
    await page.waitForTimeout(200);
    const afterWheel = await readTerminalSnapshot();
    const afterWheelLines = parseLineMarkers(afterWheel.text || '');
    report.snapshots.afterWheel = { ...afterWheel, lines: afterWheelLines };
    if (
      beforeLines.min != null &&
      afterWheelLines.min != null &&
      afterWheelLines.min < beforeLines.min
    ) {
      wheelMoved = true;
    } else if (
      beforeSnap.scrollTop != null &&
      afterWheel.scrollTop != null &&
      afterWheel.scrollTop !== beforeSnap.scrollTop
    ) {
      wheelMoved = true;
    }
  } catch (e) {
    report.issues.push('wheel baseline failed: ' + String(e && e.message ? e.message : e));
  }
  report.gesture.wheelMoved = wheelMoved;
  screenshots.push(await shot('03-after-wheel-baseline'));

  // Did touch pan reveal older lines?
  let touchScrolled = false;
  let reason = '';
  if (beforeLines.min == null || afterTouchLines.min == null) {
    reason = 'could not compare LINE markers before/after touch';
    report.issues.push(reason);
  } else if (afterTouchLines.min < beforeLines.min) {
    touchScrolled = true;
    reason = `touch revealed older lines: min ${beforeLines.min} -> ${afterTouchLines.min}`;
  } else if (
    beforeSnap.scrollTop != null &&
    afterTouchSnap.scrollTop != null &&
    afterTouchSnap.scrollTop < beforeSnap.scrollTop - 5
  ) {
    // scrollTop decreased often means scrolled up into history (depends on DOM).
    touchScrolled = true;
    reason = `touch changed scrollTop ${beforeSnap.scrollTop} -> ${afterTouchSnap.scrollTop}`;
  } else {
    touchScrolled = false;
    reason =
      'touch pan did not reveal older scrollback ' +
      `(min LINE before=${beforeLines.min} after=${afterTouchLines.min}; ` +
      `scrollTop before=${beforeSnap.scrollTop} after=${afterTouchSnap.scrollTop})`;
  }

  // Symptom: mobile touch cannot scroll terminal history.
  report.symptomPresent = !touchScrolled;
  report.okHealthy = touchScrolled && beforeLines.count > 0;
  report.symptom = {
    reason,
    touchScrolled,
    wheelMoved,
    beforeMin: beforeLines.min,
    afterTouchMin: afterTouchLines.min,
    beforeMax: beforeLines.max,
    afterTouchMax: afterTouchLines.max,
  };
  if (!touchScrolled) {
    report.issues.push(reason);
  }

  report.screenshots = screenshots;
  report.page = { title: await page.title(), url: page.url() };
  screenshots.push(await shot('04-final'));
  report.screenshots = screenshots;
} catch (err) {
  report.issues.push(String(err && err.message ? err.message : err));
  report.symptomPresent = false;
  report.okHealthy = false;
  try {
    screenshots.push(await shot('error'));
  } catch (_) {}
  report.screenshots = screenshots;
}

console.log('REPORT_JSON=' + JSON.stringify(report));
process.exit(0);
