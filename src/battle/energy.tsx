import { Fragment, type CSSProperties, type ReactNode } from 'react';
import './energy.css';

// Energy symbols (M3-1): one inline SVG glyph per type, drawn here as
// lookalikes of the TCG symbols. Placeholder shapes, the owner's to judge.

export const ENERGY_COLOR: Record<string, string> = {
  Grass: '#3a9d3f', Fire: '#e2481f', Water: '#1f86de', Lightning: '#e8b800',
  Psychic: '#9a4fc6', Fighting: '#b8611f', Darkness: '#2c3646', Metal: '#7f8d9b',
  Dragon: '#a88a1c', Colorless: '#9da3ad',
};

// The printed [X] letters in card text, and the type each one means.
const LETTER_TYPE: Record<string, string> = {
  C: 'Colorless', D: 'Darkness', F: 'Fighting', G: 'Grass', L: 'Lightning',
  M: 'Metal', N: 'Dragon', P: 'Psychic', R: 'Fire', W: 'Water',
};

// All glyphs are white, drawn on a 24x24 grid; evenodd cuts holes.
const GLYPH: Record<string, ReactNode> = {
  Grass: <path d="M4 20C3 10 9 4 20 4c1 11-5 17-13 16l5-7-6 5Z" />,
  Fire: <path d="M12 2c1 4 6 6 6 12a6 6 0 0 1-12 0c0-3 1.5-5 3-7 .5 2 1 3 2 3 1-2 0-5 1-8Z" />,
  Water: <path d="M12 2.5C9 7 6 10.5 6 15a6 6 0 0 0 12 0c0-4.500-3-8-6-12.500Z" />,
  Lightning: <path d="M14 2 5 13.500h6L10 22l9-12h-6Z" />,
  Psychic: <path fillRule="evenodd" d="M1.500 12C5 7 8.500 5.500 12 5.500S19 7 22.500 12C19 17 15.500 18.500 12 18.500S5 17 1.500 12ZM12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8Zm0 2.200a1.800 1.800 0 1 1 0 3.600 1.800 1.800 0 0 1 0-3.600Z" />,
  Fighting: <><path d="M6 10.500a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2V15a6 6 0 0 1-6 6h-1a6 6 0 0 1-6-6Z" /><circle cx="8.500" cy="7.500" r="2" /><circle cx="12" cy="6.500" r="2" /><circle cx="15.500" cy="7.500" r="2" /></>,
  Darkness: <path d="M15 2.500A9.500 9.500 0 1 0 21.500 15 7.500 7.500 0 0 1 15 2.500Z" />,
  Metal: <path fillRule="evenodd" d="m12 2 8.500 5v10L12 22l-8.500-5V7ZM12 8l-3.500 2v4l3.500 2 3.500-2v-4Z" />,
  Dragon: <path d="M12 21 3 6c3 .5 6 2.500 9 6.500C15 8.500 18 6.500 21 6Z M12 2.500l2 4h-4Z" />,
  Colorless: <path d="m12 2.500 2.900 6.200 6.600.8-4.900 4.600 1.300 6.700L12 17.500l-5.900 3.300 1.300-6.700-4.900-4.600 6.600-.8Z" />,
};

export function EnergyIcon({ type, size = 14, spent = false, testId = 'energy-icon' }: { type: string; size?: number; spent?: boolean; testId?: string }) {
  const style = { '--s': size, background: ENERGY_COLOR[type] ?? '#b9bcc8' } as CSSProperties;
  return (
    <span className="bt-energy" style={style} data-testid={testId} data-type={type} data-spent={spent} title={spent ? 'used' : type} aria-label={spent ? 'energy used' : `${type} energy`}>
      {!spent && <svg viewBox="0 0 24 24" aria-hidden="true" fill="#fff">{GLYPH[type]}</svg>}
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
