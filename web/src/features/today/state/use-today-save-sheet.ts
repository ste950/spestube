import { collectionRepository, videoRepository } from '@bootstrap/dependencies';
import { mapOk } from '@core/result/result';
import type { Collection } from '@domain/library/entities/collection';
import { makeLoadCollectionChoices } from '@domain/library/use-cases/load-collection-choices';
import { createSaveSheetStore, type SheetGroup } from '@shared/save-sheet/save-sheet-store';
import { useTodayStore } from './use-today-store';

const loadChoices = makeLoadCollectionChoices(collectionRepository, videoRepository);

const toGroup = (collection: Collection): SheetGroup => ({
  id: collection.id,
  label: collection.name,
  count: collection.videoCount,
});

// Il pannello «Salva in» di Oggi. Quando cambia l'appartenenza di un video, lo dice allo
// store di Oggi, che riempie o svuota il segnalibro della riga.
export const useTodaySaveSheet = createSaveSheetStore({
  load: async (videoId) =>
    mapOk(await loadChoices(videoId), ({ collections, selected }) => ({
      groups: collections.map(toGroup),
      selected,
    })),
  setMember: (collectionId, videoId, member) =>
    member
      ? collectionRepository.addVideo(collectionId, videoId)
      : collectionRepository.removeVideo(collectionId, videoId),
  create: async (name) =>
    mapOk(await collectionRepository.createCollection({ name, description: '' }), toGroup),
  onChange: (videoId, selected) => useTodayStore.getState().setCollections(videoId, selected),
});
