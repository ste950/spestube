import type { Channel } from '@domain/library/entities/channel';
import { AppIcon } from '@shared/components/app-icon';
import { useTranslate } from '@shared/i18n/use-translate';
import styles from './channel-list.module.css';

type ChannelListProps = {
  channels: readonly Channel[];
  onRemove: (channel: Channel) => void;
};

export function ChannelList({ channels, onRemove }: ChannelListProps) {
  const t = useTranslate();
  return (
    <ul className={styles.list}>
      {channels.map((channel) => (
        <li key={channel.id} className={styles.row}>
          <span className={styles.text}>
            <span className={styles.title}>{channel.title}</span>
            {channel.fetchFailed && (
              <span className={styles.warning}>{t('channels_fetch_error')}</span>
            )}
          </span>
          {/* L'icona da sola non dice quale canale: il nome sta nell'etichetta del bottone. */}
          <button
            type="button"
            className={styles.remove}
            aria-label={t('channels_remove', { title: channel.title })}
            onClick={() => onRemove(channel)}
          >
            <AppIcon name="close" className={styles.removeIcon} />
          </button>
        </li>
      ))}
    </ul>
  );
}
