import { useId, useState, type FormEvent } from 'react';
import { AppButton } from '@shared/components/app-button';
import { useTranslate } from '@shared/i18n/use-translate';
import styles from './add-channel-form.module.css';

type AddChannelFormProps = {
  handleLookup: boolean;
  busy: boolean;
  error: string | null;
  // Risponde `true` se l'aggiunta è riuscita: solo allora il campo si svuota.
  onAdd: (reference: string) => Promise<boolean>;
};

export function AddChannelForm({ handleLookup, busy, error, onAdd }: AddChannelFormProps) {
  const t = useTranslate();
  const fieldId = useId();
  const hintId = useId();
  const errorId = useId();
  // Il testo che si sta scrivendo è stato del campo, non della feature: nessun altro lo legge.
  const [reference, setReference] = useState('');

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (reference.trim() === '') return;
    if (await onAdd(reference)) setReference('');
  }

  const describedBy = [handleLookup ? null : hintId, error === null ? null : errorId]
    .filter((id) => id !== null)
    .join(' ');

  return (
    <form className={styles.form} onSubmit={(event) => void submit(event)}>
      <label className={styles.label} htmlFor={fieldId}>
        {t('channels_add_label')}
      </label>
      <div className={styles.controls}>
        <input
          id={fieldId}
          className={styles.field}
          type="text"
          value={reference}
          onChange={(event) => setReference(event.target.value)}
          placeholder={t(
            handleLookup ? 'channels_add_placeholder_handle' : 'channels_add_placeholder_id',
          )}
          // Un indirizzo o un id: niente maiuscola iniziale, niente correzione automatica.
          autoCapitalize="none"
          autoCorrect="off"
          autoComplete="off"
          spellCheck={false}
          enterKeyHint="done"
          aria-invalid={error !== null}
          {...(describedBy === '' ? {} : { 'aria-describedby': describedBy })}
        />
        <AppButton
          type="submit"
          label={busy ? t('channels_add_working') : t('channels_add_submit')}
          disabled={busy}
        />
      </div>
      {error !== null && (
        <p id={errorId} className={styles.error} role="alert">
          {error}
        </p>
      )}
      {!handleLookup && (
        <p id={hintId} className={styles.hint}>
          {t('channels_add_hint_no_key')}
        </p>
      )}
    </form>
  );
}
