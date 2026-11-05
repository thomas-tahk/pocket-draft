import { useEffect, useRef } from 'react';

// Focus handling for the on-screen confirms (alertdialog): focus goes to the safe button
// ([data-autofocus]) when the dialog opens, and back to where it came from when it closes.
// `returnId` names the opener when the opener is re-created as the dialog closes.
export function useDialog(returnId?: string) {
  const ref = useRef<HTMLDivElement>(null);
  // Read once: development's double effect run would otherwise take the dialog's own button as the opener.
  const openerRef = useRef<HTMLElement | null | undefined>(undefined);
  if (openerRef.current === undefined) openerRef.current = document.activeElement as HTMLElement | null;
  useEffect(() => {
    const opener = openerRef.current;
    ref.current?.querySelector<HTMLElement>('[data-autofocus]')?.focus();
    return () => {
      window.setTimeout(() => {
        // Only when focus was lost with the dialog; a new screen has already placed it.
        if (document.activeElement && document.activeElement !== document.body) return;
        const back = (returnId && document.getElementById(returnId)) || (opener?.isConnected ? opener : null);
        back?.focus();
      }, 0);
    };
  }, [returnId]);
  return ref;
}

export const onEscape = (cancel: () => void) => (e: { key: string; stopPropagation: () => void }) => {
  if (e.key === 'Escape') {
    e.stopPropagation();
    cancel();
  }
};
