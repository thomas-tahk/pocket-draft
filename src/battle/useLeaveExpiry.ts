import { useEffect, useRef } from 'react';
import { useDraftStore } from '../stores/draftStore';
import { leaveMsLeft } from '../stores/run';
import { timedOutBattle } from './leaveBattle';

// A battle left for 90 s is lost (#117). Watches the saved battle from anywhere in the app,
// and also catches one whose time ran out while the tab was closed: the first check is on load.
export function useLeaveExpiry() {
  const saved = useDraftStore((s) => s.run?.inProgress);
  const recordBattle = useDraftStore((s) => s.recordBattle);
  const expiring = useRef(false);

  useEffect(() => {
    if (!saved?.leftAt) return;
    const expire = async () => {
      if (expiring.current) return;
      expiring.current = true;
      const battle = await timedOutBattle(saved, new Date());
      expiring.current = false;
      // Returned to the battle while it was being finished: it is not lost after all.
      const now = useDraftStore.getState().run?.inProgress;
      if (now?.seed === saved.seed && now.leftAt === saved.leftAt) recordBattle(battle);
    };
    const left = leaveMsLeft(saved, Date.now()) ?? 0;
    const timer = setTimeout(() => void expire(), left);
    return () => clearTimeout(timer);
  }, [saved, recordBattle]);
}
