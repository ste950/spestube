import { Link } from 'react-router';
import { isSaved, thumbnailUrl, type Video } from '@domain/library/entities/video';
import { AppIcon } from '@shared/components/app-icon';
import { clockDuration } from '../state/today-labels';
import styles from './video-row.module.css';

type VideoRowProps = {
  video: Video;
  href: string;
  /** Il nome del bottone, con dentro il titolo: in una lista «Salva» da solo non direbbe
      di quale video. */
  saveLabel: string;
  onSave: () => void;
};

export function VideoRow({ video, href, saveLabel, onSave }: VideoRowProps) {
  return (
    <li className={styles.row}>
      {/* Il collegamento è vero: si apre in un'altra scheda e si annuncia con il titolo
          del video. Il bottone gli sta accanto e non dentro: un bottone in un link non si
          raggiunge da tastiera. */}
      <Link to={href} className={styles.link}>
        <span className={styles.thumbnail}>
          {/* La miniatura è decorativa: il titolo accanto dice già tutto, e descriverla
              vorrebbe dire far leggere due volte la stessa cosa. */}
          <img
            className={styles.image}
            src={thumbnailUrl(video.id)}
            alt=""
            loading="lazy"
            decoding="async"
            referrerPolicy="no-referrer"
          />
          {video.durationSeconds !== null && (
            <span className={styles.duration}>{clockDuration(video.durationSeconds)}</span>
          )}
        </span>
        <span className={styles.text}>
          <span className={styles.title}>{video.title}</span>
          <span className={styles.channel}>{video.channelTitle}</span>
        </span>
      </Link>
      {/* Il segnalibro pieno dice che il video è già in una raccolta; il nome del bottone lo
          dice a chi non lo vede. */}
      <button
        type="button"
        className={isSaved(video) ? styles.saveActive : styles.save}
        aria-label={saveLabel}
        onClick={onSave}
      >
        <AppIcon name="save" filled={isSaved(video)} className={styles.saveIcon} />
      </button>
    </li>
  );
}
