// Screen-reader announcements for things that change without a navigation (#121).
// <Announcer/> (mounted once, outside every screen) owns the live region; announce() writes to it.
type Listener = (message: string) => void;
let listener: Listener | null = null;

export function setAnnouncer(l: Listener | null) {
  listener = l;
}

export function announce(message: string) {
  listener?.(message);
}
