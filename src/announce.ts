// Screen-reader announcements for things that change without a navigation (#121).
// <Announcer/> (mounted once, outside every screen) owns the live region; announce() writes to it.
// A message sent before the Announcer has mounted is kept and delivered when it does.
type Listener = (message: string) => void;
let listener: Listener | null = null;
let queued: string[] = [];

export function setAnnouncer(l: Listener | null) {
  listener = l;
  if (l) {
    const waiting = queued;
    queued = [];
    for (const m of waiting) l(m);
  }
}

export function announce(message: string) {
  if (listener) listener(message);
  else queued.push(message);
}
