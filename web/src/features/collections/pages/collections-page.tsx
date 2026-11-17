import { useEffect } from 'react';
import { AppButton } from '@shared/components/app-button';
import { AppIcon } from '@shared/components/app-icon';
import { AppMessage } from '@shared/components/app-message';
import { AppPageHeader } from '@shared/components/app-page-header';
import { AppSpinner } from '@shared/components/app-spinner';
import { renderMessage } from '@shared/i18n/message';
import { useTranslate } from '@shared/i18n/use-translate';
import { CollectionFormDialog } from '../components/collection-form-dialog';
import { CollectionRow } from '../components/collection-row';
import { useCollectionsStore } from '../state/use-collections-store';
import styles from './collections-page.module.css';

const EMPTY = { name: '', description: '' };

export function CollectionsPage() {
  const t = useTranslate();
  const collections = useCollectionsStore((state) => state.collections);
  const isLoading = useCollectionsStore((state) => state.isLoading);
  const hasLoaded = useCollectionsStore((state) => state.hasLoaded);
  const error = useCollectionsStore((state) => state.error);
  const isReordering = useCollectionsStore((state) => state.isReordering);
  const isCreating = useCollectionsStore((state) => state.isCreating);
  const isSaving = useCollectionsStore((state) => state.isSaving);
  const formError = useCollectionsStore((state) => state.formError);
  const load = useCollectionsStore((state) => state.load);
  const toggleReorder = useCollectionsStore((state) => state.toggleReorder);
  const move = useCollectionsStore((state) => state.move);
  const openCreate = useCollectionsStore((state) => state.openCreate);
  const closeCreate = useCollectionsStore((state) => state.closeCreate);
  const createCollection = useCollectionsStore((state) => state.create);

  useEffect(() => {
    void load();
  }, [load]);

  return (
    <section className={styles.page} aria-busy={isLoading}>
      <title>{`${t('collections_title')} · ${t('common_app_name')}`}</title>
      <AppPageHeader title={t('collections_title')} subtitle={t('collections_subtitle')} />

      <div className={styles.toolbar}>
        <button type="button" className={styles.newButton} onClick={openCreate}>
          <AppIcon name="plus" className={styles.toolbarIcon} />
          {t('collections_new')}
        </button>
        {collections.length > 1 && (
          <AppButton
            variant="quiet"
            label={isReordering ? t('collections_reorder_done') : t('collections_reorder')}
            onClick={toggleReorder}
          />
        )}
      </div>

      {error !== null && <AppMessage tone="error" text={renderMessage(t, error)} />}
      {isLoading && !hasLoaded && <AppSpinner label={t('common_loading')} />}
      {hasLoaded && collections.length === 0 && (
        <AppMessage tone="neutral" text={t('collections_empty')} />
      )}

      {collections.length > 0 && (
        <ul className={styles.list}>
          {collections.map((collection, index) => (
            <CollectionRow
              key={collection.id}
              collection={collection}
              reordering={isReordering}
              isFirst={index === 0}
              isLast={index === collections.length - 1}
              onMove={(direction) => void move(collection.id, direction)}
            />
          ))}
        </ul>
      )}

      {isCreating && (
        <CollectionFormDialog
          title={t('collections_form_new_title')}
          initial={EMPTY}
          busy={isSaving}
          error={formError === null ? null : renderMessage(t, formError)}
          onSubmit={(fields) => void createCollection(fields)}
          onClose={closeCreate}
        />
      )}
    </section>
  );
}
