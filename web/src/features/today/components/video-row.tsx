import { Link } from 'react-router';
import { thumbnailUrl, type Video } from '@domain/library/entities/video';
import { clockDuration } from '../state/today-labels';
import styles from './video-row.module.css';

type VideoRowProps = {
  video: Video;
  href: string;
};

export function VideoRow({ video, href }: VideoRowProps) {
  return (
    <li className={styles.row}>
      {/* Tutta la riga è un collegamento vero: si apre in un'altra scheda e si annuncia
          con il titolo del video. */}
      <Link to={href} className={styles.link}>
        <span className={styles.thumbnail}>
          {/* La miniatura è decorativa: il titolo accanto dice già tutto, e descriverla
              vorrebbe dire far leggere due volte la stessa cosa. */}
          <img
            className={styles.image}
            src={thumbnailUrl(video)}
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
    </li>
  );
}
