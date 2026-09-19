// Boots the Go engine compiled to WebAssembly and routes the game's API calls
// into it, so the app runs with no server behind it.
//
// The shim patches `fetch` rather than changing `api.ts`, so every caller keeps
// using the same relative /api/* URLs and neither build knows the difference.
// Opt in with VITE_WASM=1; without it the app talks to the Go HTTP server as
// before.

declare global {
  interface Window {
    Go: new () => { importObject: WebAssembly.Imports; run(i: WebAssembly.Instance): void };
    __pdInit?: (cardsJson: string) => string;
    __pdCall?: (path: string, body?: string) => string;
  }
}

const API_PREFIX = '/api/';

function loadScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const el = document.createElement('script');
    el.src = src;
    el.onload = () => resolve();
    el.onerror = () => reject(new Error(`failed to load ${src}`));
    document.head.appendChild(el);
  });
}

function jsonResponse(body: string): Response {
  return new Response(body, {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  });
}

// installFetchShim sends /api/* at the in-page engine and leaves every other
// request — card data, images — on the real network.
function installFetchShim() {
  const realFetch = window.fetch.bind(window);
  window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.pathname : input.url;
    const path = url.startsWith('http') ? new URL(url).pathname : url;
    if (!path.startsWith(API_PREFIX) || !window.__pdCall) return realFetch(input, init);

    const body = typeof init?.body === 'string' ? init.body : undefined;
    return jsonResponse(window.__pdCall(path, body));
  };
}

/** Loads the engine and points the app's API calls at it. Throws if it can't. */
export async function bootWasmEngine(): Promise<void> {
  await loadScript(`${import.meta.env.BASE_URL}wasm_exec.js`);

  const [wasm, cards] = await Promise.all([
    WebAssembly.instantiateStreaming(
      fetch(`${import.meta.env.BASE_URL}pd.wasm`),
      new window.Go().importObject,
    ),
    fetch(`${import.meta.env.BASE_URL}data/cards.json`).then((r) => r.text()),
  ]);

  const go = new window.Go();
  // Re-instantiating against this Go's importObject keeps the runtime and the
  // module paired; instantiateStreaming above is only used to fetch and compile.
  const instance = await WebAssembly.instantiate(wasm.module, go.importObject);
  go.run(instance);

  if (!window.__pdInit) throw new Error('engine did not register __pdInit');
  const err = window.__pdInit(cards);
  if (err) throw new Error(`engine init failed: ${err}`);

  installFetchShim();
}
