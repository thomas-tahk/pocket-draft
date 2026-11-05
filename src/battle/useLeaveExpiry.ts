import { useEffect, useRef } from 'react';
import { useDraftStore } from '../stores/draftStore';
import { leaveMsLeft } from '../stores/run';
import { timedOutBattle } from './leaveBattle';

// A battle left for 90 s is lost (#117), whether it was left with the menu (leftAt) or the page
// was hidden or closed (closedAt). Watches the saved battle from anywhere in the app,
// and also catches one whose time ran out while the tab was closed: the first check is on load.
export function useLeaveExpiry(onExpired: () => void) {
  const saved = useDraftStore((s) => s.run?.inProgress);
  const recordBattle = useDraftStore((s) => s.recordBattle);
  const expiring = useRef(false);

  useEffect(() => {
    if (!saved?.leftAt && !saved?.closedAt) return;
    const expire = async () => {
      if (expiring.current) return;
      expiring.current = true;
      const unclaimed = () => {
        const now = useDraftStore.getState().run?.inProgress;
        return now?.seed === saved.seed && now.leftAt === saved.leftAt && now.closedAt === saved.closedAt;
      };
      const battle = await timedOutBattle(saved, new Date(), unclaimed);
      expiring.current = false;
      // Returned to the battle (or a resume claimed it) before the rebuild started or while it ran: not lost after all.
      if (!battle || !unclaimed()) return;
      recordBattle(battle);
      onExpired();
    };
    const left = leaveMsLeft(saved, Date.now()) ?? 0;
    const timer = setTimeout(() => void expire(), left);
    return () => clearTimeout(timer);
  }, [saved, recordBattle, onExpired]);
}

// A page that is hidden or closed mid-battle starts the same 90 s clock, so a closed tab cannot keep
// a battle alive forever. iOS often skips pagehide, so a hidden tab counts too. Coming back inside
// the 90 s stops the clock; the battle is still on screen.
export function useCloseStamp() {
  useEffect(() => {
    const stamp = () => {
      const { run, setInProgress } = useDraftStore.getState();
      const saved = run?.inProgress;
      if (saved && !saved.leftAt && !saved.closedAt) setInProgress({ ...saved, closedAt: new Date().toISOString() });
    };
    const unstamp = () => {
      const { run, setInProgress } = useDraftStore.getState();
      const saved = run?.inProgress;
      if (saved?.closedAt && (leaveMsLeft(saved, Date.now()) ?? 0) > 0) setInProgress({ ...saved, closedAt: null });
    };
    const onVisibility = () => (document.visibilityState === 'hidden' ? stamp() : unstamp());
    window.addEventListener('pagehide', stamp);
    document.addEventListener('visibilitychange', onVisibility);
    return () => {
      window.removeEventListener('pagehide', stamp);
      document.removeEventListener('visibilitychange', onVisibility);
    };
  }, []);
}
