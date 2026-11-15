import { collectionRepository } from '@bootstrap/dependencies';
import { createCollectionsStore } from './collections-store';
import { createCollectionStore } from './collection-store';

export const useCollectionsStore = createCollectionsStore(collectionRepository);

export const useCollectionStore = createCollectionStore(collectionRepository);
