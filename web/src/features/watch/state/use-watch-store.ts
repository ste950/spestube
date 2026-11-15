import { videoRepository } from '@bootstrap/dependencies';
import { createWatchStore } from './watch-store';

export const useWatchStore = createWatchStore(videoRepository);
