import { Link } from 'react-router';
import {
  durationMinutes,
  QUEUE_LIFETIME_DAYS,
  queueElapsed,
  type QueueEntry,
  type Video,
} from '@domain/library/entities/video';
import { useTranslate } from '@shared/i18n/use-translate';
import { daysLeftLabel, isExpiring } from '../state/queue-labels';
import styles from './queue-row.module.css';

export function QueueRow({
  video,
  entry,
  href,
}: {
  video: Video;
  entry: QueueEntry;
  href: string;
}) {
  const t = useTranslate();
  const minutes = durationMinutes(video);
  const expiring = isExpiring(entry.daysLeft);

  return (
    <li className={expiring ? styles.rowExpiring : styles.row}>
      <Link to={href} className={styles.title}>
        {video.title}
      </Link>
      <p className={styles.meta}>
        {minutes === null
          ? video.channelTitle
          : `${video.channelTitle} · ${t('common_duration_minutes', { minutes })}`}
      </p>
      <p className={styles.reason}>
        <span className={styles.reasonPrefix}>{t('queue_reason_prefix')}</span> {entry.reason}
      </p>
      <div className={styles.expiry}>
        {/* La barra ripete a colpo d'occhio ciò che il testo accanto dice per intero: per
            chi ascolta è solo rumore, quindi è nascosta. */}
        <span className={styles.track} aria-hidden="true">
          <span
            className={styles.trackFill}
            style={{ width: `${Math.round(queueElapsed(entry, QUEUE_LIFETIME_DAYS) * 100)}%` }}
          />
        </span>
        <span className={styles.daysLeft}>
          {t(daysLeftLabel(entry.daysLeft), { days: entry.daysLeft })}
        </span>
      </div>
    </li>
  );
}
