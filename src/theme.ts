// Light / dark theme (#131). 'auto' follows the phone (prefers-color-scheme); 'light' and 'dark'
// override it by setting data-theme on <html>, which tokens.css reads. The choice is kept in
// localStorage; storage can be missing or blocked (private mode), so every access is guarded.

export type ThemeChoice = 'auto' | 'light' | 'dark';
export const THEME_KEY = 'pocket-draft-theme';
export const THEME_CHOICES: readonly ThemeChoice[] = ['auto', 'light', 'dark'];

const isChoice = (v: unknown): v is ThemeChoice => THEME_CHOICES.includes(v as ThemeChoice);

export function readTheme(): ThemeChoice {
  try {
    const v = localStorage.getItem(THEME_KEY);
    return isChoice(v) ? v : 'auto';
  } catch {
    return 'auto';
  }
}

export function applyTheme(choice: ThemeChoice) {
  const root = document.documentElement;
  if (choice === 'auto') root.removeAttribute('data-theme');
  else root.setAttribute('data-theme', choice);
}

export function saveTheme(choice: ThemeChoice) {
  applyTheme(choice);
  try {
    if (choice === 'auto') localStorage.removeItem(THEME_KEY);
    else localStorage.setItem(THEME_KEY, choice);
  } catch {
    // Not persisted; the choice still holds for this page.
  }
}
