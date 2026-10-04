import { useEffect, useState } from 'react';
import Notification from '../Notification/Notification';

/** Global host for `eh:achievement` events from syncAchievements. */
export function AchievementToast() {
  const [message, setMessage] = useState<string | null>(null);

  useEffect(() => {
    const onAchievement = (ev: Event) => {
      const detail = (ev as CustomEvent<{ message?: string }>).detail;
      if (detail?.message) {
        setMessage(detail.message);
      }
    };
    window.addEventListener('eh:achievement', onAchievement);
    return () => window.removeEventListener('eh:achievement', onAchievement);
  }, []);

  if (!message) return null;

  return (
    <Notification type="success" message={message} onClose={() => setMessage(null)} autoClose={4000} />
  );
}
