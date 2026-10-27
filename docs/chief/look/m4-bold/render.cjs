// Screenshots mockup.html for each direction and screen at 390×844 (issue #131).
// Run from the repo root: node docs/chief/look/m4-bold/render.cjs
const { chromium } = require('@playwright/test');
const path = require('path');
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2 });
  const file = 'file://' + path.join(__dirname, 'mockup.html');
  for (const dir of ['burst', 'holo', 'type'])
    for (const screen of ['draft', 'run']) {
      await page.goto(`${file}?dir=${dir}&screen=${screen}`);
      await page.screenshot({ path: path.join(__dirname, `${dir}-${screen}-390.png`) });
    }
  // One overview image: a column per direction, draft above run.
  const names = { burst: 'A · Pack Burst', holo: 'B · Holo Night', type: 'C · Type Colour' };
  const cols = Object.entries(names).map(([d, n]) => `<div><h2>${n}</h2><img src="${d}-draft-390.png"><img src="${d}-run-390.png"></div>`).join('');
  await page.setViewportSize({ width: 1250, height: 1000 });
  await page.goto(file);
  await page.setContent(`<base href="${file}"><style>body{margin:0;background:#111;font:900 28px system-ui;color:#fff;display:flex;gap:20px;padding:20px}
    img{display:block;width:390px;margin-top:12px;border-radius:12px}h2{margin:0;font-size:28px}</style>${cols}`);
  await page.waitForFunction(() => [...document.images].every((i) => i.complete));
  await page.screenshot({ path: path.join(__dirname, 'overview.png'), fullPage: true });
  await browser.close();
})();
