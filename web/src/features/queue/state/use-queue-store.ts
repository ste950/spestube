import { videoRepository } from '@bootstrap/dependencies';
import { createQueueStore } from './queue-store';

export const useQueueStore = createQueueStore(videoRepository);
