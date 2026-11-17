import { useId, useState, type FormEvent } from 'react';
import {
  COLLECTION_DESCRIPTION_MAX,
  COLLECTION_NAME_MAX,
} from '@domain/library/entities/collection';
import type { CollectionFields } from '@domain/library/repositories/collection-repository';
import { AppButton } from '@shared/components/app-button';
import { AppDialog } from '@shared/components/app-dialog';
import fields from '@shared/components/app-field.module.css';
import { useTranslate } from '@shared/i18n/use-translate';
import styles from './collection-form-dialog.module.css';

type CollectionFormDialogProps = {
  title: string;
  initial: CollectionFields;
  busy: boolean;
  error: string | null;
  onSubmit: (fields: CollectionFields) => void;
  onClose: () => void;
};

// Il form di una raccolta, per crearla e per modificarla: un nome e una descrizione
// facoltativa. La validazione vera è del server; qui si evita solo l'invio di un nome vuoto.
export function CollectionFormDialog({
  title,
  initial,
  busy,
  error,
  onSubmit,
  onClose,
}: CollectionFormDialogProps) {
  const t = useTranslate();
  const nameId = useId();
  const descriptionId = useId();
  const hintId = useId();
  const errorId = useId();
  const [name, setName] = useState(initial.name);
  const [description, setDescription] = useState(initial.description);

  function submit(event: FormEvent) {
    event.preventDefault();
    onSubmit({ name, description });
  }

  return (
    <AppDialog title={title} busy={busy} onClose={onClose}>
      <form className={styles.form} onSubmit={submit}>
        <label className={fields.label} htmlFor={nameId}>
          {t('collections_form_name')}
        </label>
        <input
          id={nameId}
          className={fields.field}
          value={name}
          onChange={(event) => setName(event.target.value)}
          placeholder={t('collections_form_name_placeholder')}
          maxLength={COLLECTION_NAME_MAX}
          autoComplete="off"
          // La finestra modale porta il focus sul primo campo che lo chiede.
          autoFocus
          aria-invalid={error !== null}
          {...(error === null ? {} : { 'aria-describedby': errorId })}
        />
        {error !== null && (
          <p id={errorId} className={fields.error} role="alert">
            {error}
          </p>
        )}

        <label className={fields.label} htmlFor={descriptionId}>
          {t('collections_form_description')}
        </label>
        <textarea
          id={descriptionId}
          className={fields.field}
          value={description}
          onChange={(event) => setDescription(event.target.value)}
          rows={3}
          maxLength={COLLECTION_DESCRIPTION_MAX}
          aria-describedby={hintId}
        />
        <p id={hintId} className={fields.hint}>
          {t('collections_form_description_hint')}
        </p>

        <div className={fields.actions}>
          <AppButton label={t('common_cancel')} variant="quiet" onClick={onClose} disabled={busy} />
          <AppButton
            type="submit"
            label={busy ? t('common_saving') : t('common_save')}
            disabled={busy}
          />
        </div>
      </form>
    </AppDialog>
  );
}
