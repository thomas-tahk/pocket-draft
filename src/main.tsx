import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { App } from './App';
import { Announcer } from './components/Announcer';
import { bootWasmEngine } from './game/wasmBoot';
import { applyTheme, readTheme } from './theme';
import './tokens.css';
import './index.css';
import './screens.css';
import './look.css';

// index.html's inline script sets it before the first paint; this is the backstop.
applyTheme(readTheme());

const root = createRoot(document.getElementById('root')!);

function render() {
  root.render(
    <StrictMode>
      <App />
      <Announcer />
    </StrictMode>,
  );
}

// In the serverless build the Go engine runs in the page, so it has to be up
// before anything can start a game. Everywhere else the Go HTTP server answers.
if (import.meta.env.VITE_WASM === '1') {
  root.render(<p style={{ padding: 24, font: '16px system-ui' }}>Loading the game engine…</p>);
  bootWasmEngine().then(render, (err: Error) => {
    root.render(<p style={{ padding: 24, font: '16px system-ui' }}>Engine failed to load: {err.message}</p>);
  });
} else {
  render();
}
