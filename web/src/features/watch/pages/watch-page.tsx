import { useEffect } from 'react';
import { Link } from 'react-router';
import { durationMinutes } from '@domain/library/entities/video';
import { AppButton } from '@shared/components/app-button';
import { AppIcon } from '@shared/components/app-icon';
import { AppMessage } from '@shared/components/app-message';
import { AppSpinner } from '@shared/components/app-spinner';
import { renderMessage } from '@shared/i18n/message';
import { useTranslate } from '@shared/i18n/use-translate';
import { StopScreen } from '../components/stop-screen';
import { YoutubePlayer } from '../components/youtube-player';
import { useWatchStore } from '../state/use-watch-store';
import styles from './watch-page.module.css';

export function WatchPage({ videoId }: { videoId: string }) {
  const t = useTranslate();
  const video = useWatchStore((state) => state.video);
  const isLoading = useWatchStore((state) => state.isLoading);
  const error = useWatchStore((state) => state.error);
  const hasEnded = useWatchStore((state) => state.hasEnded);
  const isMarking = useWatchStore((state) => state.isMarking);
  const load = useWatchStore((state) => state.load);
  const end = useWatchStore((state) => state.end);
  const markWatched = useWatchStore((state) => state.markWatched);

  useEffect(() => {
    void load(videoId);
  }, [load, videoId]);

  const minutes = video === null ? null : durationMinutes(video);
  const watched = video?.status === 'watched';

  return (
    <article className={styles.page} aria-busy={isLoading}>
      <title>{`${video?.title ?? t('common_loading')} · ${t('common_app_name')}`}</title>
      <Link to="/" className={styles.back}>
        <AppIcon name="back" className={styles.backIcon} />
        {t('watch_back')}
      </Link>

      {isLoading && <AppSpinner label={t('common_loading')} />}
      {error !== null && <AppMessage tone="error" text={renderMessage(t, error)} />}

      {video !== null && (
        <>
          {hasEnded ? (
            <StopScreen title={t('watch_end_title')} text={t('watch_end_text')} />
          ) : (
            <YoutubePlayer
              videoId={video.id}
              frameTitle={t('watch_player_title', { title: video.title })}
              loadingLabel={t('watch_player_loading')}
              errorLabel={t('watch_player_error')}
              onEnded={end}
            />
          )}

          <header className={styles.heading}>
            <h1 className={styles.title}>{video.title}</h1>
            <p className={styles.meta}>
              {minutes === null
                ? video.channelTitle
                : `${video.channelTitle} · ${t('common_duration_minutes', { minutes })}`}
            </p>
          </header>

          {video.queue !== null && (
            <section className={styles.reason}>
              <h2 className={styles.reasonLabel}>{t('watch_reason_label')}</h2>
              <p className={styles.reasonText}>{video.queue.reason}</p>
            </section>
          )}

          <div className={styles.actions}>
            <AppButton
              label={watched ? t('watch_watched') : t('watch_mark_watched')}
              onClick={() => void markWatched()}
              disabled={watched || isMarking}
            />
            <Link to="/" className={styles.secondary}>
              {t('common_back_to_today')}
            </Link>
          </div>
        </>
      )}
    </article>
  );
}
