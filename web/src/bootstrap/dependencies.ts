import { createLocalThemeRepository } from '@data/appearance/repositories/local-theme-repository';
import { createLocalThemeSource } from '@data/appearance/sources/local-theme-source';
import {
  createRemoteChannelRepository,
  createRemoteCollectionRepository,
  createRemoteVideoRepository,
} from '@data/library/repositories/remote-library-repositories';
import { createHttpLibrarySource } from '@data/library/sources/http-library-source';
import { createRemoteSettingsRepository } from '@data/settings/repositories/remote-settings-repository';
import { createHttpSettingsSource } from '@data/settings/sources/http-settings-source';
import type { ThemeRepository } from '@domain/appearance/repositories/theme-repository';
import type { ChannelRepository } from '@domain/library/repositories/channel-repository';
import type { CollectionRepository } from '@domain/library/repositories/collection-repository';
import type { VideoRepository } from '@domain/library/repositories/video-repository';
import type { SettingsRepository } from '@domain/settings/repositories/settings-repository';
import { apiBaseUrl } from './config';

const librarySource = createHttpLibrarySource(apiBaseUrl);

// L'annotazione è tutto il lavoro di un contenitore: è qui che TypeScript verifica la
// conformità, ed è qui che il tipo esportato diventa il contratto invece dell'implementazione.
export const channelRepository: ChannelRepository = createRemoteChannelRepository(librarySource);

export const videoRepository: VideoRepository = createRemoteVideoRepository(librarySource);

export const collectionRepository: CollectionRepository =
  createRemoteCollectionRepository(librarySource);

export const settingsRepository: SettingsRepository = createRemoteSettingsRepository(
  createHttpSettingsSource(apiBaseUrl),
);

export const themeRepository: ThemeRepository =
  createLocalThemeRepository(createLocalThemeSource());
