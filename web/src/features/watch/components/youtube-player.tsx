import { useEffect, useRef, useState } from 'react';
import { loadYouTubeApi, PLAYER_HOST, type YouTubePlayer } from './youtube-iframe-api';
import styles from './youtube-player.module.css';

type YoutubePlayerProps = {
  videoId: string;
  /** Il nome accessibile dell'iframe: è ciò che il lettore di schermo annuncia entrandoci. */
  frameTitle: string;
  loadingLabel: string;
  errorLabel: string;
  onEnded: () => void;
};

export function YoutubePlayer({
  videoId,
  frameTitle,
  loadingLabel,
  errorLabel,
  onEnded,
}: YoutubePlayerProps) {
  const frame = useRef<HTMLDivElement>(null);
  const [status, setStatus] = useState<'loading' | 'ready' | 'failed'>('loading');

  useEffect(() => {
    const host = frame.current;
    if (host === null) return;

    // L'API sostituisce l'elemento che riceve con un iframe. Gliene si dà uno creato a
    // mano, così React non si trova a gestire un nodo che qualcun altro ha rimpiazzato.
    const mount = document.createElement('div');
    host.append(mount);

    let cancelled = false;
    let player: YouTubePlayer | null = null;

    loadYouTubeApi()
      .then((api) => {
        if (cancelled) return;
        player = new api.Player(mount, {
          videoId,
          host: PLAYER_HOST,
          width: '100%',
          height: '100%',
          playerVars: {
            // Mai autoplay: il video parte quando lo decide chi guarda.
            autoplay: 0,
            // Limita i suggerimenti finali al canale del video. Non si vedono comunque: a
            // fine video il player viene tolto.
            rel: 0,
            // Su iPhone resta nella pagina invece di aprirsi a tutto schermo da solo.
            playsinline: 1,
            origin: window.location.origin,
          },
          events: {
            onReady: (event) => {
              event.target.getIframe().title = frameTitle;
              setStatus('ready');
            },
            onStateChange: (event) => {
              if (event.data === api.PlayerState.ENDED) onEnded();
            },
            onError: () => setStatus('failed'),
          },
        });
      })
      .catch(() => {
        if (!cancelled) setStatus('failed');
      });

    return () => {
      cancelled = true;
      player?.destroy();
      host.replaceChildren();
    };
  }, [videoId, frameTitle, onEnded]);

  return (
    <div className={styles.player}>
      <div ref={frame} className={styles.frame} />
      {status === 'loading' && (
        <p className={styles.status} role="status">
          {loadingLabel}…
        </p>
      )}
      {status === 'failed' && (
        <p className={styles.status} role="alert">
          {errorLabel}
        </p>
      )}
    </div>
  );
}
