import { useEffect, useState } from 'react';
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
  return (
    <div className="run-confirm" data-testid="leave-countdown">
      <p>You left your battle. It counts as a loss in <strong data-testid="leave-seconds">{seconds}</strong> s.</p>
      <button className="run-btn run-btn-primary" disabled={seconds <= 0} onClick={onReturn}>Return to battle</button>
    </div>
  );
}
