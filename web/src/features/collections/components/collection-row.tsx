import { Link } from 'react-router';
import type { Collection } from '@domain/library/entities/collection';
import { thumbnailUrl } from '@domain/library/entities/video';
import { AppIcon } from '@shared/components/app-icon';
import { useTranslate } from '@shared/i18n/use-translate';
import { countLabel } from '../state/collection-labels';
import styles from './collection-row.module.css';

type CollectionRowProps = {
  collection: Collection;
  reordering: boolean;
  isFirst: boolean;
  isLast: boolean;
  onMove: (direction: 'up' | 'down') => void;
};

export function CollectionRow({
  collection,
  reordering,
  isFirst,
  isLast,
  onMove,
}: CollectionRowProps) {
  const t = useTranslate();
  return (
    <li className={styles.row}>
      <Link to={`/raccolte/${collection.id}`} className={styles.link}>
        <span className={styles.cover}>
          {collection.coverVideoId === null ? (
            <AppIcon name="collections" className={styles.coverIcon} />
          ) : (
            // Decorativa: il nome accanto dice già quale raccolta è.
            <img
              className={styles.image}
              src={thumbnailUrl(collection.coverVideoId)}
              alt=""
              loading="lazy"
              decoding="async"
              referrerPolicy="no-referrer"
            />
          )}
        </span>
        <span className={styles.text}>
          <span className={styles.name}>{collection.name}</span>
          <span className={styles.count}>
            {t(countLabel(collection.videoCount), { count: collection.videoCount })}
          </span>
        </span>
      </Link>
      {reordering && (
        <span className={styles.moves}>
          {/* Frecce e non trascinamento: si usano con un dito, con la tastiera e con un
              lettore di schermo, e su un telefono non si confondono con lo scorrimento. */}
          <button
            type="button"
            className={styles.move}
            aria-label={t('collections_move_up', { name: collection.name })}
            disabled={isFirst}
            onClick={() => onMove('up')}
          >
            <AppIcon name="up" className={styles.moveIcon} />
          </button>
          <button
            type="button"
            className={styles.move}
            aria-label={t('collections_move_down', { name: collection.name })}
            disabled={isLast}
            onClick={() => onMove('down')}
          >
            <AppIcon name="down" className={styles.moveIcon} />
          </button>
        </span>
      )}
    </li>
  );
}
