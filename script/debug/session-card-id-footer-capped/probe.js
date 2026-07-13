// Session card bottom footer: truncated session id under prompt title.
//
// Symptom (unfixed): cards with a human prompt still show .session-item-id
// using shortSessionId mid-ellipsis (e.g. "brainstorm…3-083040"), which looks
// capped/junk at the bottom of the card.
//
// Healthy: no such footer when prompt is present, OR id is not mid-ellipsis-capped
// for long human-readable session ids.
//
//   playwright-debug run probe.js <baseURL> <outDir> [token]
// Prints REPORT_JSON=... ; always exit 0 (orchestrator interprets).

const baseURL = (process.argv[3] || 'http://127.0.0.1:8192').replace(/\/$/, '');
const outDir = process.argv[4] || '/tmp/session-card-id-footer-capped';
const apiToken = process.argv[5] || '';

const fs = require('fs');
const path = require('path');
fs.mkdirSync(outDir, { recursive: true });

const SEED_SESSION_ID = 'brainstorm-add-terminal-color-skill-20260713-083040';
const SEED_PROMPT = '/brainstorm add terminal color skill';

// Product shortSessionId: len>20 → first10 + … + last8
function expectedShortSessionId(id) {
  if (id.length <= 20) return id;
  return `${id.slice(0, 10)}…${id.slice(-8)}`;
}

const expectedCapped = expectedShortSessionId(SEED_SESSION_ID);

async function shot(name) {
  const file = path.join(outDir, `${name}.png`);
  await page.screenshot({ path: file, fullPage: false });
  return file;
}

if (apiToken) {
  await page.addInitScript((tok) => {
    try {
      localStorage.setItem('agent-run-token', tok);
    } catch (_) {}
  }, apiToken);
}

await page.setViewportSize({ width: 390, height: 844 });
await page.goto(`${baseURL}/`, { waitUntil: 'networkidle', timeout: 45000 });
await page.waitForSelector('[data-testid="session-list"], [data-testid="empty-state"]', {
  timeout: 20000,
});
await page.waitForTimeout(500);

// Prefer the seeded long-id card; fall back to any prompt card with .session-item-id
const cardProbe = await page.evaluate(
  ({ seedId, seedPrompt, expectedCapped }) => {
    const items = Array.from(document.querySelectorAll('[data-testid="session-item"]'));
    const rows = items.map((el) => {
      const preview = (el.querySelector('[data-testid="session-preview"]')?.textContent || '').trim();
      const idEl = el.querySelector('.session-item-id');
      const idText = idEl ? (idEl.textContent || '').trim() : null;
      const idTitle = idEl?.getAttribute('title') || null;
      const idVisible = Boolean(
        idEl &&
          idEl.getBoundingClientRect().height > 2 &&
          getComputedStyle(idEl).visibility !== 'hidden' &&
          getComputedStyle(idEl).display !== 'none',
      );
      return {
        preview,
        idText,
        idTitle,
        idVisible,
        hasIdFooter: Boolean(idEl),
        href: el.getAttribute('href') || '',
      };
    });

    let target =
      rows.find((r) => r.href.includes(encodeURIComponent(seedId)) || r.idTitle === seedId) ||
      rows.find((r) => r.preview.includes('brainstorm') && r.hasIdFooter) ||
      rows.find((r) => r.hasIdFooter && r.preview && r.idText && r.idText.includes('…'));

    // Mid-ellipsis capped id: contains … and is shorter than full title
    const isMidEllipsisCap = (text, full) => {
      if (!text) return false;
      const hasEllipsis = text.includes('…') || text.includes('...');
      if (!hasEllipsis) return false;
      if (full && text === full) return false;
      if (full && text.length < full.length) return true;
      // shortSessionId shape: roughly 10 + ellipsis + 8
      return text.length <= 22 && hasEllipsis;
    };

    const symptomCards = rows.filter(
      (r) =>
        r.hasIdFooter &&
        r.idVisible &&
        r.preview &&
        r.idText &&
        isMidEllipsisCap(r.idText, r.idTitle || ''),
    );

    return {
      rowCount: rows.length,
      target,
      symptomCards,
      expectedCapped,
      seedId,
      seedPrompt,
      anyIdFooter: rows.some((r) => r.hasIdFooter),
      sample: rows.slice(0, 6),
    };
  },
  { seedId: SEED_SESSION_ID, seedPrompt: SEED_PROMPT, expectedCapped },
);

const screenshots = [await shot('01-home')];

// If seed card found, crop-ish full page already; try highlight by scrolling to it
if (cardProbe.target?.href) {
  await page.evaluate((href) => {
    const el = document.querySelector(`a[href="${href}"]`);
    el?.scrollIntoView({ block: 'center' });
  }, cardProbe.target.href);
  await page.waitForTimeout(200);
  screenshots.push(await shot('02-target-card'));
}

const target = cardProbe.target;
const expectedMatch =
  target &&
  target.idText &&
  (target.idText === expectedCapped ||
    target.idText.replace(/\.\.\./g, '…') === expectedCapped);

// Symptom: prompt card shows mid-ellipsis short session id footer (user's "bottom capped")
const symptomPresent =
  Boolean(target?.hasIdFooter && target?.idVisible && target?.idText) &&
  (expectedMatch ||
    (target.idText.includes('…') &&
      target.idTitle &&
      target.idText.length < target.idTitle.length) ||
    cardProbe.symptomCards.length > 0);

const reasons = [];
if (symptomPresent) {
  reasons.push(
    `session card bottom shows capped session id footer: text=${JSON.stringify(target?.idText)} title=${JSON.stringify(target?.idTitle)}`,
  );
  if (expectedMatch) {
    reasons.push(`matches shortSessionId(seed)=${JSON.stringify(expectedCapped)}`);
  }
  if (cardProbe.symptomCards.length > 1) {
    reasons.push(`${cardProbe.symptomCards.length} cards have mid-ellipsis id footers`);
  }
}

// Healthy: no visible mid-ellipsis id footer on prompt cards
const okHealthy =
  !symptomPresent &&
  cardProbe.symptomCards.length === 0 &&
  cardProbe.rowCount > 0;

const report = {
  baseURL,
  symptomPresent,
  okHealthy,
  reasons,
  issues: [],
  symptom: {
    target,
    expectedCapped,
    seedSessionId: SEED_SESSION_ID,
    symptomCardCount: cardProbe.symptomCards.length,
  },
  snapshots: cardProbe,
  screenshots,
};

fs.writeFileSync(path.join(outDir, 'probe-report.json'), JSON.stringify(report, null, 2));
console.log('REPORT_JSON=' + JSON.stringify(report));
process.exit(0);
