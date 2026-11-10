// Puts focus on the screen's h1 (no ring on it): where focus goes when the control that had it is gone.
export function focusHeading(root: ParentNode = document) {
  const h = root.querySelector<HTMLElement>('h1');
  if (!h) return;
  h.tabIndex = -1;
  h.focus({ preventScroll: true });
}
