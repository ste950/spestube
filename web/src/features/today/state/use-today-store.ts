import { settingsRepository, videoRepository } from '@bootstrap/dependencies';
import { setNavBadge } from '@shared/navigation/nav-badges';
import { createTodayStore } from './today-store';

// Oggi è la sezione alla radice: il numero dei suoi video va accanto a quella voce della barra.
export const useTodayStore = createTodayStore(videoRepository, settingsRepository, (count) =>
  setNavBadge('/', count),
);
