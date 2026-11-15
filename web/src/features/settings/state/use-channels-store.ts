import { channelRepository, settingsRepository } from '@bootstrap/dependencies';
import { createChannelsStore } from './channels-store';

export const useChannelsStore = createChannelsStore(channelRepository, settingsRepository);
