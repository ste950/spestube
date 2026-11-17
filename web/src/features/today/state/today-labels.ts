export type NextRelease = {
  readonly day: 'today_end_day_today' | 'today_end_day_tomorrow';
  readonly time: string;
};

// Il prossimo arrivo è sempre entro ventiquattro ore: o più tardi oggi, o domani.
export function nextReleaseLabel(nextReleaseAt: Date, now: Date): NextRelease {
  const sameDay = nextReleaseAt.toDateString() === now.toDateString();
  return {
    day: sameDay ? 'today_end_day_today' : 'today_end_day_tomorrow',
    time: new Intl.DateTimeFormat('it-IT', { hour: 'numeric', minute: '2-digit' }).format(
      nextReleaseAt,
    ),
  };
}

// «18:04» per la targhetta sulla miniatura, con le ore solo quando servono.
export function clockDuration(totalSeconds: number): string {
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = String(Math.floor(totalSeconds % 60)).padStart(2, '0');
  return hours > 0
    ? `${hours}:${String(minutes).padStart(2, '0')}:${seconds}`
    : `${minutes}:${seconds}`;
}
