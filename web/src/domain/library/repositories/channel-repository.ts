import type { Result } from '@core/result/result';
import type { Channel } from '../entities/channel';

export interface ChannelRepository {
  listChannels(): Promise<Result<readonly Channel[]>>;
  // `reference` è ciò che l'utente ha scritto: un link, un @handle o un id. Capirlo è
  // mestiere del server.
  addChannel(reference: string): Promise<Result<Channel>>;
  removeChannel(channelId: string): Promise<Result<void>>;
}
