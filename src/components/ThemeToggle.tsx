import { useState } from 'react';
import { readTheme, saveTheme, THEME_CHOICES, type ThemeChoice } from '../theme';

const LABEL: Record<ThemeChoice, string> = { auto: 'Auto', light: 'Light', dark: 'Dark' };

// Auto / Light / Dark. Self-contained (its own styles in look.css, gold for the chosen option, the
// surrounding text colour for the rest), so it can sit in any header or panel.
export function ThemeToggle() {
  const [choice, setChoice] = useState<ThemeChoice>(readTheme);
  return (
    <div className="theme-toggle" role="group" aria-labelledby="theme-toggle-label" data-testid="theme-toggle">
      <span id="theme-toggle-label" className="theme-toggle-label">Theme</span>
      <span className="theme-toggle-options">
        {THEME_CHOICES.map((c) => (
          <button
            key={c}
            type="button"
            aria-pressed={choice === c}
            onClick={() => {
              setChoice(c);
              saveTheme(c);
            }}
          >
            {LABEL[c]}
          </button>
        ))}
      </span>
    </div>
  );
}
