import { test, expect } from '@playwright/test';
import { existsSync, readFileSync } from 'node:fs';
import { dirname, join, relative } from 'node:path';

// 1.29: the old battle screen (the `?play` GameView) is gone now that the new
// board is reachable from the draft. Nothing the app bundles may import it.

const OLD_SCREEN = [
  'src/game/GameView.tsx',
  'src/game/ActionBar.tsx',
  'src/game/GameCard.tsx',
  'src/game/Hand.tsx',
  'src/game/Log.tsx',
  'src/game/PlayerSide.tsx',
  'src/game/diff.ts',
];

const SPEC = /(?:from\s+|import\s*\(?\s*)['"](\.{1,2}\/[^'"]+)['"]/g;

function resolve(fromFile: string, spec: string): string | null {
  const base = join(dirname(fromFile), spec);
  const tries = [base, `${base}.ts`, `${base}.tsx`, join(base, 'index.ts'), join(base, 'index.tsx')];
  return tries.find((p) => existsSync(p) && !p.endsWith('/')) ?? null;
}

// Every local file reachable by import from the app's entry point.
function bundled(entry: string): Set<string> {
  const seen = new Set<string>();
  const queue = [entry];
  while (queue.length) {
    const file = queue.pop()!;
    if (seen.has(file) || !/\.tsx?$/.test(file)) continue;
    seen.add(file);
    for (const m of readFileSync(file, 'utf8').matchAll(SPEC)) {
      const next = resolve(file, m[1]);
      if (next) queue.push(next);
    }
  }
  return new Set([...seen].map((f) => relative('.', f)));
}

test('the build imports nothing from the old battle screen', () => {
  const files = bundled('src/main.tsx');
  expect(files.has('src/battle/BattleScreen.tsx')).toBe(true);
  expect(OLD_SCREEN.filter((f) => files.has(f))).toEqual([]);
});

test('?play no longer opens the old screen; the draft app loads instead', async ({ page }) => {
  await page.goto('/?play');
  await expect(page.getByRole('button', { name: /start/i }).first()).toBeVisible();
  await expect(page.getByText('Your hand')).toHaveCount(0);
});
