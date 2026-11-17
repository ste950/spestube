// La parte dell'IFrame Player API di YouTube che l'app usa. È l'unico modo in cui un video
// viene riprodotto: il player ufficiale, incorporato, e nient'altro.
export type YouTubePlayer = {
  destroy(): void;
  getIframe(): HTMLIFrameElement;
};

type PlayerOptions = {
  videoId: string;
  host: string;
  width: string;
  height: string;
  playerVars: Record<string, string | number>;
  events: {
    onReady: (event: { target: YouTubePlayer }) => void;
    onStateChange: (event: { data: number }) => void;
    onError: () => void;
  };
};

export type YouTubeApi = {
  Player: new (element: HTMLElement, options: PlayerOptions) => YouTubePlayer;
  PlayerState: { ENDED: number };
};

declare global {
  interface Window {
    YT?: YouTubeApi;
    onYouTubeIframeAPIReady?: () => void;
  }
}

// Il dominio senza cookie di YouTube: finché non si preme play, l'incorporamento non
// associa la visita a un profilo.
export const PLAYER_HOST = 'https://www.youtube-nocookie.com';

const API_SCRIPT = 'https://www.youtube.com/iframe_api';

let pending: Promise<YouTubeApi> | null = null;

// Carica lo script dell'API una volta sola, la prima volta che serve un player.
export function loadYouTubeApi(): Promise<YouTubeApi> {
  if (window.YT?.Player !== undefined) return Promise.resolve(window.YT);
  pending ??= new Promise<YouTubeApi>((resolve, reject) => {
    window.onYouTubeIframeAPIReady = () => {
      if (window.YT !== undefined) resolve(window.YT);
    };
    const script = document.createElement('script');
    script.src = API_SCRIPT;
    script.async = true;
    script.onerror = () => {
      // Un tentativo fallito non resta in memoria: aprendo un altro video si riprova.
      pending = null;
      script.remove();
      reject(new Error('lo script del player di YouTube non si è caricato'));
    };
    document.head.append(script);
  });
  return pending;
}
