import { useEffect } from 'react';
import { AppMessage } from '@shared/components/app-message';
import { AppPageHeader } from '@shared/components/app-page-header';
import { AppSpinner } from '@shared/components/app-spinner';
import { renderMessage } from '@shared/i18n/message';
import { useTranslate } from '@shared/i18n/use-translate';
import { QueueRow } from '../components/queue-row';
import { useQueueStore } from '../state/use-queue-store';
import styles from './queue-page.module.css';

export function QueuePage() {
  const t = useTranslate();
  const videos = useQueueStore((state) => state.videos);
  const isLoading = useQueueStore((state) => state.isLoading);
  const hasLoaded = useQueueStore((state) => state.hasLoaded);
  const error = useQueueStore((state) => state.error);
  const load = useQueueStore((state) => state.load);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <section className={styles.page} aria-busy={isLoading}>
      <title>{`${t('queue_title')} · ${t('common_app_name')}`}</title>
      <AppPageHeader title={t('queue_title')} subtitle={t('queue_subtitle')} />

      {error !== null && <AppMessage tone="error" text={renderMessage(t, error)} />}
      {isLoading && !hasLoaded && <AppSpinner label={t('common_loading')} />}
      {hasLoaded && error === null && videos.length === 0 && (
        <AppMessage tone="neutral" text={t('queue_empty')} />
      )}

      <ul className={styles.list}>
        {videos.map(
          (video) =>
            video.queue !== null && (
              <QueueRow
                key={video.id}
                video={video}
                entry={video.queue}
                href={`/guarda/${encodeURIComponent(video.id)}`}
              />
            ),
        )}
      </ul>
    </section>
  );
}
