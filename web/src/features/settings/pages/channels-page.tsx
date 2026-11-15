import { useEffect } from 'react';
import type { Channel } from '@domain/library/entities/channel';
import { AppMessage } from '@shared/components/app-message';
import { AppPageHeader } from '@shared/components/app-page-header';
import { AppSpinner } from '@shared/components/app-spinner';
import { renderMessage } from '@shared/i18n/message';
import { useTranslate } from '@shared/i18n/use-translate';
import { AddChannelForm } from '../components/add-channel-form';
import { ChannelList } from '../components/channel-list';
import { useChannelsStore } from '../state/use-channels-store';
import styles from './channels-page.module.css';

export function ChannelsPage() {
  const t = useTranslate();
  const channels = useChannelsStore((state) => state.channels);
  const handleLookup = useChannelsStore((state) => state.handleLookup);
  const isLoading = useChannelsStore((state) => state.isLoading);
  const hasLoaded = useChannelsStore((state) => state.hasLoaded);
  const error = useChannelsStore((state) => state.error);
  const isAdding = useChannelsStore((state) => state.isAdding);
  const addError = useChannelsStore((state) => state.addError);
  const notice = useChannelsStore((state) => state.notice);
  const load = useChannelsStore((state) => state.load);
  const add = useChannelsStore((state) => state.add);
  const remove = useChannelsStore((state) => state.remove);

  useEffect(() => {
    void load();
  }, [load]);

  function confirmRemove(channel: Channel) {
    // Rimuovere un canale cancella anche i suoi video, e non si torna indietro: si chiede
    // conferma con la finestra del sistema, che è accessibile senza doverla costruire.
    if (window.confirm(t('channels_remove_confirm', { title: channel.title }))) {
      void remove(channel.id);
    }
  }

  return (
    <section className={styles.page} aria-busy={isLoading}>
      <title>{`${t('channels_title')} · ${t('common_app_name')}`}</title>
      <AppPageHeader title={t('channels_title')} subtitle={t('channels_subtitle')} />

      {error !== null && <AppMessage tone="error" text={renderMessage(t, error)} />}

      <section className={styles.block} aria-labelledby="canali-seguiti">
        <div className={styles.blockHeading}>
          <h2 id="canali-seguiti" className={styles.blockTitle}>
            {t('channels_list_title')}
          </h2>
          {hasLoaded && <span className={styles.count}>{channels.length}</span>}
        </div>
        {isLoading && !hasLoaded && <AppSpinner label={t('common_loading')} />}
        {hasLoaded && channels.length === 0 && (
          <AppMessage tone="neutral" text={t('channels_empty')} />
        )}
        {channels.length > 0 && <ChannelList channels={channels} onRemove={confirmRemove} />}
      </section>

      <AddChannelForm
        handleLookup={handleLookup}
        busy={isAdding}
        error={addError === null ? null : renderMessage(t, addError)}
        onAdd={add}
      />
      {notice !== null && <AppMessage tone="success" text={renderMessage(t, notice)} />}
    </section>
  );
}
