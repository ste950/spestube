import { useEffect } from 'react';
import { Link } from 'react-router';
import { durationMinutes } from '@domain/library/entities/video';
import { AppButton } from '@shared/components/app-button';
import { AppIcon } from '@shared/components/app-icon';
import { AppMessage } from '@shared/components/app-message';
import { AppSpinner } from '@shared/components/app-spinner';
import { renderMessage } from '@shared/i18n/message';
import { SaveSheet } from '@shared/save-sheet/save-sheet';
import { useTranslate } from '@shared/i18n/use-translate';
import { StopScreen } from '../components/stop-screen';
import { YoutubePlayer } from '../components/youtube-player';
import { useWatchSaveSheet } from '../state/use-watch-save-sheet';
import { useWatchStore } from '../state/use-watch-store';
import styles from './watch-page.module.css';

type WatchPageProps = {
  videoId: string;
  /** La raccolta da cui si è aperto il video: il ritorno porta lì e non a Oggi. */
  fromCollection: number | null;
};

export function WatchPage({ videoId, fromCollection }: WatchPageProps) {
  const t = useTranslate();
  const video = useWatchStore((state) => state.video);
  const isLoading = useWatchStore((state) => state.isLoading);
  const error = useWatchStore((state) => state.error);
  const hasEnded = useWatchStore((state) => state.hasEnded);
  const isMarking = useWatchStore((state) => state.isMarking);
  const load = useWatchStore((state) => state.load);
  const end = useWatchStore((state) => state.end);
  const markWatched = useWatchStore((state) => state.markWatched);
  const openSaveSheet = useWatchSaveSheet((state) => state.open);

  useEffect(() => {
    void load(videoId);
  }, [load, videoId]);

  const minutes = video === null ? null : durationMinutes(video);
  const watched = video?.status === 'watched';
  const savedIn = video?.collectionIds.length ?? 0;
  const back =
    fromCollection === null
      ? { to: '/', label: t('watch_back') }
      : { to: `/raccolte/${fromCollection}`, label: t('watch_back_collection') };

  return (
    <article className={styles.page} aria-busy={isLoading}>
      <title>{`${video?.title ?? t('common_loading')} · ${t('common_app_name')}`}</title>
      <Link to={back.to} className={styles.back}>
        <AppIcon name="back" className={styles.backIcon} />
        {back.label}
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

          <div className={styles.actions}>
            <AppButton
              label={watched ? t('watch_watched') : t('watch_mark_watched')}
              onClick={() => void markWatched()}
              disabled={watched || isMarking}
            />
            <button
              type="button"
              className={savedIn > 0 ? styles.saveActive : styles.save}
              onClick={() => void openSaveSheet(video.id, video.title)}
            >
              <AppIcon name="save" filled={savedIn > 0} className={styles.saveIcon} />
              {savedIn === 0
                ? t('watch_save')
                : t(savedIn === 1 ? 'watch_saved_one' : 'watch_saved_many', { count: savedIn })}
            </button>
            <Link to={back.to} className={styles.secondary}>
              {fromCollection === null ? t('common_back_to_today') : t('watch_back_to_collection')}
            </Link>
          </div>
        </>
      )}
      <SaveSheet store={useWatchSaveSheet} />
    </article>
  );
}
