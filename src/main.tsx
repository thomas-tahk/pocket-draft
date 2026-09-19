import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { App } from './App';
import { bootWasmEngine } from './game/wasmBoot';
import './index.css';

const root = createRoot(document.getElementById('root')!);

function render() {
  root.render(
    <StrictMode>
      <App />
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
