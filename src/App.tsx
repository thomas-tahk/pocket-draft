import { useEffect, useState } from 'react';
import { loadCardPool, type CardPool } from './data';

export function App() {
  const [pool, setPool] = useState<CardPool | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadCardPool().then(setPool).catch((e) => setError(String(e)));
  }, []);

  if (error) return <main style={{ padding: 24 }}>Error: {error}</main>;
  if (!pool) return <main style={{ padding: 24 }}>Loading card data…</main>;

  return (
    <main style={{ padding: 24, maxWidth: 720, margin: '0 auto' }}>
      <h1>Pocket Draft</h1>
      <p>v0.1 data layer loaded.</p>
      <ul>
        <li>{pool.allCards.length} raw cards</li>
        <li>{pool.draftableCards.length} draftable (base-art, no cosmetics)</li>
        <li>{pool.sets.length} sets, {Object.keys(pool.packs).length} packs</li>
      </ul>
    </main>
  );
}
