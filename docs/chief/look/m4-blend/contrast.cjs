// WCAG 2 contrast ratios for the key colour pairs of the M4 blended look (src/tokens.css).
// Run from the repo root: node docs/chief/look/m4-blend/contrast.cjs
// Keep the hex values here in step with tokens.css; the README table is this script's output.
const lum = (hex) => {
  const h = hex.replace('#', '');
  const [r, g, b] = [0, 2, 4]
    .map((i) => parseInt(h.slice(i, i + 2), 16) / 255)
    .map((c) => (c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4));
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
};
const ratio = (a, b) => {
  const [hi, lo] = [lum(a), lum(b)].sort((x, y) => y - x);
  return (hi + 0.05) / (lo + 0.05);
};

// [theme, pair, fg, bg, needed]  needed: 4.5 body text, 3 large text / UI / focus ring.
const PAIRS = [
  ['light', 'Body text on panel', '#0d1a33', '#ffffff', 4.5],
  ['light', 'Muted text on panel', '#3d4c6b', '#ffffff', 4.5],
  ['light', 'Muted text on inner row', '#3d4c6b', '#eef3fc', 4.5],
  ['light', 'Section label / button text on panel', '#0047a3', '#ffffff', 4.5],
  ['light', 'Button text on button hover', '#0047a3', '#dfeaff', 4.5],
  ['light', 'Text on page blue', '#ffffff', '#0b63c9', 4.5],
  ['light', 'Muted text on page blue', '#e3eeff', '#0b63c9', 4.5],
  ['light', 'Text on header band (dark end)', '#ffffff', '#003c99', 4.5],
  ['light', 'Text on header band (light end)', '#ffffff', '#0057c2', 4.5],
  ['light', 'Main button text on gold (top)', '#2a1600', '#ffd23f', 4.5],
  ['light', 'Main button text on gold (bottom)', '#2a1600', '#ffa800', 4.5],
  ['light', 'Quit run text on soft red', '#b0101a', '#ffe5e5', 4.5],
  ['light', 'Win mark letter on green', '#ffffff', '#0b7d3b', 4.5],
  ['light', 'Loss mark letter on red', '#ffffff', '#d61f2b', 4.5],
  ['light', 'Loss word on inner row', '#c4141c', '#eef3fc', 4.5],
  ['light', 'Win word on panel', '#0e7a3c', '#ffffff', 4.5],
  ['light', 'Outline ink on page blue (UI)', '#1a1030', '#0b63c9', 3],
  ['light', 'Focus ring on page (white)', '#ffffff', '#0b63c9', 3],
  ['light', 'Focus ring on band (gold)', '#ffd23f', '#0057c2', 3],
  ['light', 'Focus ring on panel (ink)', '#1a1030', '#ffffff', 3],
  ['light', 'Theme toggle: chosen option on gold', '#2a1600', '#ffd23f', 4.5],
  ['light', 'Theme toggle: other options on band', '#ffffff', '#003c99', 4.5],
  ['light', 'Pack segment (gold) on band (UI)', '#ffd23f', '#0057c2', 3],
  ['dark', 'Body text on panel', '#f3f0ff', '#1c1452', 4.5],
  ['dark', 'Body text on page', '#f3f0ff', '#140b3d', 4.5],
  ['dark', 'Muted text on panel', '#c9bff5', '#1c1452', 4.5],
  ['dark', 'Muted text on page (darkest)', '#c9bff5', '#090620', 4.5],
  ['dark', 'Section label on panel', '#b9a8ff', '#1c1452', 4.5],
  ['dark', 'Text on inner row', '#f3f0ff', '#251a5e', 4.5],
  ['dark', 'Text on header band (dark end)', '#ffffff', '#2b1680', 4.5],
  ['dark', 'Text on header band (light end)', '#ffffff', '#5326c4', 4.5],
  ['dark', 'Button text on button', '#ffffff', '#2a1f6e', 4.5],
  ['dark', 'Main button text on gold (bottom)', '#2a1600', '#ffa800', 4.5],
  ['dark', 'Quit run text on soft red', '#ffb3c5', '#3d0f2e', 4.5],
  ['dark', 'Loss word on inner row', '#ff8fab', '#251a5e', 4.5],
  ['dark', 'Win word on panel', '#5ff0a0', '#1c1452', 4.5],
  ['dark', 'Win mark letter on green', '#ffffff', '#0b7d3b', 4.5],
  ['dark', 'Loss mark letter on red', '#ffffff', '#d61f2b', 4.5],
  ['dark', 'Outline rim on page (UI)', '#8f7bff', '#140b3d', 3],
  ['dark', 'Focus ring (gold) on panel', '#ffd23f', '#1c1452', 3],
  ['dark', 'Focus ring (gold) on page', '#ffd23f', '#090620', 3],
];

const md = process.argv.includes('--md');
if (md) console.log('| Theme | Pair | Colours | Ratio | Needs |\n|---|---|---|---|---|');
let bad = 0;
for (const [theme, name, fg, bg, need] of PAIRS) {
  const r = ratio(fg, bg);
  if (r < need) bad++;
  if (md) console.log(`| ${theme} | ${name} | \`${fg}\` on \`${bg}\` | ${r.toFixed(2)}:1 | ${need}:1 ${r >= need ? 'pass' : '**FAIL**'} |`);
  else console.log(`${r >= need ? 'ok  ' : 'FAIL'} ${theme.padEnd(5)} ${name.padEnd(40)} ${fg} on ${bg}  ${r.toFixed(2)} (needs ${need})`);
}
process.exitCode = bad ? 1 : 0;
