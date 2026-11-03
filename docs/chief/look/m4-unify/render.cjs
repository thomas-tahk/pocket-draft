// Screenshots mockup.html for each direction and screen at 390×844. Run from repo root: node docs/chief/look/m4-unify/render.cjs
const { chromium } = require('playwright');
const path = require('path');
(async () => {
  const b = await chromium.launch();
  const p = await b.newPage({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2 });
  const file = 'file://' + path.join(__dirname, 'mockup.html');
  const dirs = ['board', 'cream', 'night'], screens = ['home', 'draft', 'run', 'battle', 'replay'];
  for (const d of dirs) for (const s of screens) { await p.goto(`${file}?dir=${d}&screen=${s}`); await p.screenshot({ path: path.join(__dirname, `${d}-${s}-390.png`) }); }
  for (const d of dirs) {
    const row = screens.map((s) => `<img src="${d}-${s}-390.png">`).join('');
    await p.setViewportSize({ width: 2000, height: 900 });
    await p.setContent(`<base href="${file}"><style>body{margin:0;background:#111;display:flex;gap:14px;padding:14px}img{width:390px;border-radius:10px}</style>${row}`);
    await p.waitForFunction(() => [...document.images].every((i) => i.complete));
    await p.screenshot({ path: path.join(__dirname, `${d}-all.png`), fullPage: true });
  }
  await b.close();
})();
