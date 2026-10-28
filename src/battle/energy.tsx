import { Fragment, type CSSProperties, type ReactNode } from 'react';
import './energy.css';
import { ENERGY_PATH } from './energyGlyphs';

// Energy symbols (M3-1): the official TCG symbol per type (energyGlyphs.ts),
// white on the type's colored disc.

// The colours live in src/tokens.css (--type-*).
export const ENERGY_COLOR: Record<string, string> = Object.fromEntries(
  ['Grass', 'Fire', 'Water', 'Lightning', 'Psychic', 'Fighting', 'Darkness', 'Metal', 'Dragon', 'Colorless'].map((t) => [t, `var(--type-${t.toLowerCase()})`]),
);

// The printed [X] letters in card text, and the type each one means.
const LETTER_TYPE: Record<string, string> = {
  C: 'Colorless', D: 'Darkness', F: 'Fighting', G: 'Grass', L: 'Lightning',
  M: 'Metal', N: 'Dragon', P: 'Psychic', R: 'Fire', W: 'Water',
};

export function EnergyIcon({ type, size = 14, spent = false, testId = 'energy-icon' }: { type: string; size?: number; spent?: boolean; testId?: string }) {
  const style = { '--s': size, background: ENERGY_COLOR[type] ?? 'var(--type-spent)' } as CSSProperties;
  return (
    <span className="bt-energy" style={style} data-testid={testId} data-type={type} data-spent={spent} title={spent ? 'used' : type} aria-label={spent ? 'energy used' : `${type} energy`}>
      {!spent && <svg viewBox="0 0 24 24" aria-hidden="true" fill="#fff"><path d={ENERGY_PATH[type]} /></svg>}
    </span>
  );
}

// Card text with [W]-style letters (and "Fire energy" in the move log) drawn as icons.
const TEXT_ENERGY = /\[([CDFGLMNPRW])\]|\b(Grass|Fire|Water|Lightning|Psychic|Fighting|Darkness|Metal|Dragon|Colorless)(?= [Ee]nergy)/g;

export function EnergyText({ text, size = 14 }: { text: string; size?: number }) {
  const parts: ReactNode[] = [];
  let last = 0;
  for (const m of text.matchAll(TEXT_ENERGY)) {
    parts.push(text.slice(last, m.index));
    parts.push(<EnergyIcon key={m.index} type={m[1] ? LETTER_TYPE[m[1]] : m[2]} size={size} testId="energy-text-icon" />);
    last = m.index + m[0].length;
  }
  parts.push(text.slice(last));
  return <>{parts.map((p, i) => <Fragment key={i}>{p}</Fragment>)}</>;
}

// "Fire ×3  Water ×2" as icons with counts, for the deck-tally lines.
export function TypeTally({ entries }: { entries: [string, number][] }) {
  return (
    <>
      {entries.map(([type, n]) => (
        <span key={type} className="bt-tally" data-testid="type-tally">
          <EnergyIcon type={type} size={14} testId="tally-icon" /> ×{n}
        </span>
      ))}
    </>
  );
}
