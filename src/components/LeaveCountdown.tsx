import { useEffect, useRef, useState } from 'react';
import { focusHeading } from '../focus';
import { announce } from '../announce';
import { leaveMsLeft, type InProgressBattle } from '../stores/run';

// Shown on the run screen while a left battle can still be returned to (#117).
// The loss itself is recorded by useLeaveExpiry; this only counts down.
export function LeaveCountdown({ battle, onReturn }: { battle: InProgressBattle; onReturn: () => void }) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const tick = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(tick);
  }, []);
  // The count ticks every second: announce it once, when the countdown starts.
  useEffect(() => {
    announce(`You left your battle. It counts as a loss in ${Math.ceil((leaveMsLeft(battle, Date.now()) ?? 0) / 1000)} s.`);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  const seconds = Math.ceil((leaveMsLeft(battle, now) ?? 0) / 1000);
  // The Return button goes away when time runs out (disabled, then the whole countdown): focus must not
  // fall to the page. It goes to the heading.
  const box = useRef<HTMLDivElement>(null);
  const hadFocus = useRef(false);
  useEffect(() => {
    return () => {
      if (!hadFocus.current) return;
      window.setTimeout(() => {
        const run = document.querySelector('[data-screen="run"]'); // not when it left for the battle
        if (run && (!document.activeElement || document.activeElement === document.body)) focusHeading(run);
      }, 0);
    };
  }, []);
  useEffect(() => {
    if (seconds <= 0 && hadFocus.current) {
      hadFocus.current = false;
      focusHeading(document.querySelector('[data-screen="run"]') ?? document);
    }
  }, [seconds]);
  return (
    <div
      ref={box}
      className="run-confirm"
      data-testid="leave-countdown"
      onFocus={() => (hadFocus.current = true)}
      onBlur={(e) => {
        if (!e.currentTarget.contains(e.relatedTarget as Node | null) && document.activeElement !== document.body) hadFocus.current = false;
      }}
    >
      <p>You left your battle. It counts as a loss in <strong data-testid="leave-seconds">{seconds}</strong> s.</p>
      <button className="run-btn run-btn-primary" disabled={seconds <= 0} onClick={onReturn}>Return to battle</button>
    </div>
  );
}
