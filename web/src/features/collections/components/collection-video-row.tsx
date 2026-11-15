import { Link } from 'react-router';
import type { CollectionVideo } from '@domain/library/entities/collection';
import { thumbnailUrl } from '@domain/library/entities/video';
import { AppIcon } from '@shared/components/app-icon';
import styles from './collection-video-row.module.css';

type CollectionVideoRowProps = {
  video: CollectionVideo;
  href: string;
  removeLabel: string;
  onRemove: () => void;
};

export function CollectionVideoRow({
  video,
  href,
  removeLabel,
  onRemove,
}: CollectionVideoRowProps) {
  return (
    <li className={styles.row}>
      <Link to={href} className={styles.link}>
        <span className={styles.thumbnail}>
          <img
            className={styles.image}
            src={thumbnailUrl(video.id)}
            alt=""
            loading="lazy"
            decoding="async"
            referrerPolicy="no-referrer"
          />
        </span>
        <span className={styles.text}>
          <span className={styles.title}>{video.title}</span>
          <span className={styles.channel}>{video.channelTitle}</span>
        </span>
      </Link>
      <button type="button" className={styles.remove} aria-label={removeLabel} onClick={onRemove}>
        <AppIcon name="close" className={styles.removeIcon} />
      </button>
    </li>
  );
}
