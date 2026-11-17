import { useEffect } from 'react';
import { Link, useNavigate } from 'react-router';
import { AppButton } from '@shared/components/app-button';
import { AppIcon } from '@shared/components/app-icon';
import { AppMessage } from '@shared/components/app-message';
import { AppSpinner } from '@shared/components/app-spinner';
import { renderMessage } from '@shared/i18n/message';
import { useTranslate } from '@shared/i18n/use-translate';
import { CollectionFormDialog } from '../components/collection-form-dialog';
import { CollectionVideoRow } from '../components/collection-video-row';
import { countLabel } from '../state/collection-labels';
import { useCollectionStore } from '../state/use-collections-store';
import styles from './collection-page.module.css';

export function CollectionPage({ collectionId }: { collectionId: number }) {
  const t = useTranslate();
  const navigate = useNavigate();
  const detail = useCollectionStore((state) => state.detail);
  const isLoading = useCollectionStore((state) => state.isLoading);
  const error = useCollectionStore((state) => state.error);
  const isEditing = useCollectionStore((state) => state.isEditing);
  const isSaving = useCollectionStore((state) => state.isSaving);
  const formError = useCollectionStore((state) => state.formError);
  const load = useCollectionStore((state) => state.load);
  const openEdit = useCollectionStore((state) => state.openEdit);
  const closeEdit = useCollectionStore((state) => state.closeEdit);
  const update = useCollectionStore((state) => state.update);
  const remove = useCollectionStore((state) => state.remove);
  const removeVideo = useCollectionStore((state) => state.removeVideo);

  useEffect(() => {
    void load(collectionId);
  }, [load, collectionId]);

  async function confirmDelete(name: string) {
    // Eliminare una raccolta non si annulla: si chiede conferma con la finestra del sistema,
    // che è accessibile senza doverla costruire. I video restano, sparisce solo la raccolta.
    if (!window.confirm(t('collection_delete_confirm', { name }))) return;
    if (await remove()) void navigate('/raccolte', { replace: true });
  }

  const collection = detail?.collection;

  return (
    <section className={styles.page} aria-busy={isLoading}>
      <title>{`${collection?.name ?? t('collections_title')} · ${t('common_app_name')}`}</title>
      <Link to="/raccolte" className={styles.back}>
        <AppIcon name="back" className={styles.backIcon} />
        {t('collection_back')}
      </Link>

      {isLoading && <AppSpinner label={t('common_loading')} />}
      {error !== null && <AppMessage tone="error" text={renderMessage(t, error)} />}

      {detail !== null && collection !== undefined && (
        <>
          <header className={styles.header}>
            <h1 className={styles.title}>{collection.name}</h1>
            {collection.description !== '' && (
              <p className={styles.description}>{collection.description}</p>
            )}
            <p className={styles.count}>
              {t(countLabel(collection.videoCount), { count: collection.videoCount })}
            </p>
          </header>

          <div className={styles.actions}>
            <AppButton variant="quiet" label={t('collection_edit')} onClick={openEdit} />
            <button
              type="button"
              className={styles.delete}
              onClick={() => void confirmDelete(collection.name)}
            >
              {t('collection_delete')}
            </button>
          </div>

          {detail.videos.length === 0 ? (
            <AppMessage tone="neutral" text={t('collection_empty')} />
          ) : (
            <ul className={styles.list}>
              {detail.videos.map((video) => (
                <CollectionVideoRow
                  key={video.id}
                  video={video}
                  href={`/guarda/${encodeURIComponent(video.id)}?raccolta=${collection.id}`}
                  removeLabel={t('collection_remove', {
                    name: collection.name,
                    title: video.title,
                  })}
                  onRemove={() => void removeVideo(video.id)}
                />
              ))}
            </ul>
          )}

          {isEditing && (
            <CollectionFormDialog
              title={t('collections_form_edit_title')}
              initial={{ name: collection.name, description: collection.description }}
              busy={isSaving}
              error={formError === null ? null : renderMessage(t, formError)}
              onSubmit={(fields) => void update(fields)}
              onClose={closeEdit}
            />
          )}
        </>
      )}
    </section>
  );
}
