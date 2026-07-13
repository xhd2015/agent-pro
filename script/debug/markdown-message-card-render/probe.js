// Playwright inspect for markdown rendering on thinking + assistant cards.
// Invoked via:
//   playwright-debug run probe.js <baseURL> <outDir> [token] [sessionPath]
//
// sessionPath default: /sessions/layout-md-render
//
// Prints one REPORT_JSON=... line. process.exit(0) always from playwright
// (orchestrator interprets report.symptomPresent / report.okHealthy).

const baseURL = (process.argv[3] || 'http://127.0.0.1:8192').replace(/\/$/, '');
const outDir = process.argv[4] || '/tmp/markdown-message-card-render';
const apiToken = process.argv[5] || '';
const sessionPath = process.argv[6] || '/sessions/layout-md-render';

const fs = require('fs');
const path = require('path');

fs.mkdirSync(outDir, { recursive: true });

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

/**
 * Healthy ("looks good") criteria for seeded markdown:
 * - assistant response has real <strong> for **bold**
 * - assistant has <pre> and/or <code> for fenced/inline code
 * - assistant textContent must not still show raw ** for the seeded bold markers
 * - thinking progress card body has markdown structure when seed includes markers
 *   (strong or code), not only truncated plain string with **
 */
function evaluateQuality(snapshot) {
  const issues = [];
  const assistants = snapshot.assistants || [];
  const thinks = snapshot.thinks || [];

  if (assistants.length === 0) {
    issues.push('no assistant-message cards found');
  }

  let anyStrong = false;
  let anyPre = false;
  let anyCode = false;
  let anyLiteralStars = false;
  let anyLiteralFence = false;

  for (const a of assistants) {
    if (a.hasStrong) anyStrong = true;
    if (a.hasPre) anyPre = true;
    if (a.hasCode) anyCode = true;
    if (a.literalStars) anyLiteralStars = true;
    if (a.literalFence) anyLiteralFence = true;
  }

  if (!anyStrong) {
    issues.push('assistant: missing <strong> for bold markdown (looks like plain text)');
  }
  if (!anyPre && !anyCode) {
    issues.push('assistant: missing <pre>/<code> for fenced or inline code');
  }
  if (anyLiteralStars) {
    issues.push('assistant: still shows raw ** markers in textContent (not rendered)');
  }
  if (anyLiteralFence) {
    issues.push('assistant: still shows raw ``` fence markers in textContent');
  }

  // Thinking cards: if any body still contains ** or ` as raw-only (no strong/code), flag.
  let thinkMarkdownOK = thinks.length === 0;
  if (thinks.length > 0) {
    const withMarkers = thinks.filter(
      (t) => (t.text || '').includes('**') || /`[^`]+`/.test(t.text || ''),
    );
    if (withMarkers.length === 0) {
      // seed may use plain think text — require at least non-empty body
      thinkMarkdownOK = thinks.some((t) => (t.text || '').trim().length > 0);
      if (!thinkMarkdownOK) {
        issues.push('thinking: progress-card body empty');
      }
    } else {
      thinkMarkdownOK = withMarkers.some((t) => t.hasStrong || t.hasCode);
      if (!thinkMarkdownOK) {
        issues.push(
          'thinking: body has markdown markers but no <strong>/<code> (plain pre-wrap only)',
        );
      }
    }
  }

  const okHealthy =
    issues.length === 0 &&
    anyStrong &&
    (anyPre || anyCode) &&
    !anyLiteralStars &&
    !anyLiteralFence &&
    thinkMarkdownOK;

  // Symptom for bug-repro: markdown not rendered (plain text markers).
  const symptomPresent =
    assistants.length > 0 &&
    (!anyStrong || anyLiteralStars || anyLiteralFence || (!anyPre && !anyCode));

  let reason = '';
  if (symptomPresent) {
    const parts = [];
    if (!anyStrong) parts.push('no <strong>');
    if (!anyPre && !anyCode) parts.push('no <pre>/<code>');
    if (anyLiteralStars) parts.push('literal **');
    if (anyLiteralFence) parts.push('literal ```');
    reason = 'markdown not rendered on assistant response: ' + parts.join(', ');
  }

  return {
    okHealthy,
    symptomPresent,
    reason,
    issues,
    signals: {
      anyStrong,
      anyPre,
      anyCode,
      anyLiteralStars,
      anyLiteralFence,
      thinkMarkdownOK,
      assistantCount: assistants.length,
      thinkCount: thinks.length,
    },
  };
}

async function captureSnapshot() {
  return page.evaluate(() => {
    const assistants = Array.from(
      document.querySelectorAll('[data-testid="assistant-message"]'),
    ).map((el) => {
      const text = (el.textContent || '').trim();
      const html = el.innerHTML || '';
      return {
        text: text.slice(0, 500),
        html: html.slice(0, 800),
        hasStrong: !!el.querySelector('strong, b'),
        hasPre: !!el.querySelector('pre'),
        hasCode: !!el.querySelector('code'),
        hasMarkdownRoot: !!el.querySelector('[data-testid="markdown-body"], .markdown-body'),
        literalStars: text.includes('**'),
        literalFence: text.includes('```'),
        whiteSpace: getComputedStyle(el).whiteSpace,
        fontSize: getComputedStyle(el).fontSize,
      };
    });

    const thinks = Array.from(document.querySelectorAll('[data-testid="progress-card"]'))
      .filter((card) => {
        const label = (card.querySelector('.progress-card-label')?.textContent || '').trim();
        return /thinking/i.test(label);
      })
      .map((card) => {
        const body = card.querySelector('.progress-card-body') || card;
        const text = (body.textContent || '').trim();
        const html = body.innerHTML || '';
        return {
          text: text.slice(0, 400),
          html: html.slice(0, 600),
          hasStrong: !!body.querySelector('strong, b'),
          hasCode: !!body.querySelector('code'),
          hasPre: !!body.querySelector('pre'),
          hasMarkdownRoot: !!body.querySelector(
            '[data-testid="markdown-body"], .markdown-body',
          ),
          maxHeight: getComputedStyle(body).maxHeight,
          overflow: getComputedStyle(body).overflow,
        };
      });

    const users = Array.from(
      document.querySelectorAll('[data-testid="message-item-user"] .message-body'),
    ).map((el) => (el.textContent || '').trim().slice(0, 200));

    return {
      url: location.href,
      assistants,
      thinks,
      users,
      progressCardCount: document.querySelectorAll('[data-testid="progress-card"]').length,
    };
  });
}

const screenshots = [];
const report = {
  baseURL,
  sessionPath,
  symptomPresent: false,
  okHealthy: false,
  issues: [],
  symptom: {},
  snapshot: {},
  screenshots: [],
  page: {},
};

try {
  await page.setViewportSize({ width: 390, height: 844 });
  await seedTokenIfNeeded();

  const url = baseURL + sessionPath;
  await page.goto(url, { waitUntil: 'networkidle', timeout: 45000 });
  await page.waitForSelector('[data-testid="message-list"], [data-testid="chat-active"]', {
    timeout: 20000,
  });
  // Wait for at least one message bubble (seeded session).
  await page
    .waitForSelector(
      '[data-testid="assistant-message"], [data-testid="message-item-user"]',
      { timeout: 20000 },
    )
    .catch(() => {});

  screenshots.push(await shot('01-session-loaded'));

  // Prefer focusing message list for a tighter visual artifact.
  const list = page.locator('[data-testid="message-list"]');
  if ((await list.count()) > 0) {
    const file = path.join(outDir, '02-message-list.png');
    await list.screenshot({ path: file }).catch(() => {});
    if (fs.existsSync(file)) screenshots.push(file);
  }

  // Screenshot first assistant card if present.
  const asst = page.locator('[data-testid="assistant-message"]').first();
  if ((await asst.count()) > 0) {
    const file = path.join(outDir, '03-assistant-card.png');
    await asst.screenshot({ path: file }).catch(() => {});
    if (fs.existsSync(file)) screenshots.push(file);
  }

  const thinkCard = page.locator('[data-testid="progress-card"]').filter({
    has: page.locator('.progress-card-label', { hasText: /thinking/i }),
  }).first();
  if ((await thinkCard.count()) > 0) {
    const file = path.join(outDir, '04-thinking-card.png');
    await thinkCard.screenshot({ path: file }).catch(() => {});
    if (fs.existsSync(file)) screenshots.push(file);
  }

  const snapshot = await captureSnapshot();
  const quality = evaluateQuality(snapshot);

  report.snapshot = snapshot;
  report.symptomPresent = quality.symptomPresent;
  report.okHealthy = quality.okHealthy;
  report.issues = quality.issues;
  report.symptom = {
    reason: quality.reason,
    signals: quality.signals,
  };
  report.screenshots = screenshots;
  report.page = {
    title: await page.title(),
    url: page.url(),
  };

  // Extra full-page after evaluation for visual review.
  screenshots.push(await shot('05-final'));
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
