import { useEffect, useState } from 'react';
import { setAnnouncer } from '../announce';

// The one polite live region. Visually hidden; a repeated message still re-announces
// because a trailing non-breaking space alternates.
export function Announcer() {
  const [message, setMessage] = useState('');
  useEffect(() => {
    let flip = false;
    setAnnouncer((m) => {
      flip = !flip;
      setMessage(flip ? m : `${m} `);
    });
    return () => setAnnouncer(null);
  }, []);
  return (
    <div role="status" aria-live="polite" aria-atomic="true" className="sr-only" data-testid="announcer">
      {message}
    </div>
  );
}
