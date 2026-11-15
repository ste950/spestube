import { useEffect } from 'react';
import { AppMessage } from '@shared/components/app-message';
import { AppSpinner } from '@shared/components/app-spinner';
import { AppVisuallyHidden } from '@shared/components/app-visually-hidden';
import { renderMessage } from '@shared/i18n/message';
import { useTranslate } from '@shared/i18n/use-translate';
import { DevTools } from '../components/dev-tools';
import { FeedEnd } from '../components/feed-end';
import { VideoRow } from '../components/video-row';
import { nextReleaseLabel } from '../state/today-labels';
import { useTodayStore } from '../state/use-today-store';
import styles from './today-page.module.css';

export function TodayPage() {
  const t = useTranslate();
  const videos = useTodayStore((state) => state.videos);
  const nextReleaseAt = useTodayStore((state) => state.nextReleaseAt);
  const devTools = useTodayStore((state) => state.devTools);
  const isLoading = useTodayStore((state) => state.isLoading);
  const error = useTodayStore((state) => state.error);
  const notice = useTodayStore((state) => state.notice);
  const isReleasing = useTodayStore((state) => state.isReleasing);
  const load = useTodayStore((state) => state.load);
  const releaseNow = useTodayStore((state) => state.releaseNow);

  useEffect(() => {
    void load();
  }, [load]);

  const now = new Date();
  const ready = nextReleaseAt !== null && error === null;
  const next = nextReleaseAt === null ? null : nextReleaseLabel(nextReleaseAt, now);

  return (
    <section className={styles.page} aria-busy={isLoading}>
      <title>{`${t('today_title')} · ${t('common_app_name')}`}</title>
      {/* Nessuna intestazione a video: la sezione accesa nella barra dice già dove si è, e
          accanto porta il numero dei video. Il titolo resta per chi ascolta. */}
      <AppVisuallyHidden as="h1">{t('today_title')}</AppVisuallyHidden>

      {error !== null && <AppMessage tone="error" text={renderMessage(t, error)} />}
      {isLoading && nextReleaseAt === null && <AppSpinner label={t('common_loading')} />}

      {/* La lista è tutta qui: niente «carica altri», niente pagina successiva. */}
      {ready && videos.length > 0 && (
        <ul className={styles.list}>
          {videos.map((video) => (
            <VideoRow
              key={video.id}
              video={video}
              href={`/guarda/${encodeURIComponent(video.id)}`}
            />
          ))}
        </ul>
      )}

      {ready && next !== null && (
        <FeedEnd
          title={t('today_end_title')}
          text={t('today_end_text', { day: t(next.day), time: next.time })}
        />
      )}

      {notice !== null && <AppMessage tone="success" text={renderMessage(t, notice)} />}
      {devTools && (
        <DevTools
          title={t('today_dev_title')}
          text={t('today_dev_text')}
          actionLabel={t('today_dev_release')}
          busy={isReleasing}
          onRelease={() => void releaseNow()}
        />
      )}
    </section>
  );
}
