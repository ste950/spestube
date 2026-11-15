import { useId, useState, type FormEvent } from 'react';
import { AppButton } from '../components/app-button';
import { AppDialog } from '../components/app-dialog';
import { AppMessage } from '../components/app-message';
import { AppSpinner } from '../components/app-spinner';
import fields from '../components/app-field.module.css';
import { renderMessage } from '../i18n/message';
import { useTranslate } from '../i18n/use-translate';
import type { SaveSheetStore } from './save-sheet-store';
import styles from './save-sheet.module.css';

// Il pannello «Salva in…»: una casella per raccolta, che si spunta e si toglie subito, e
// in fondo il campo per crearne una nuova. Si rende solo mentre c'è un elemento aperto.
export function SaveSheet({ store }: { store: SaveSheetStore }) {
  const t = useTranslate();
  const itemId = store((state) => state.itemId);
  const itemTitle = store((state) => state.itemTitle);
  const groups = store((state) => state.groups);
  const selected = store((state) => state.selected);
  const pending = store((state) => state.pending);
  const isLoading = store((state) => state.isLoading);
  const isCreating = store((state) => state.isCreating);
  const error = store((state) => state.error);
  const close = store((state) => state.close);
  const toggle = store((state) => state.toggle);
  const create = store((state) => state.create);
  const [name, setName] = useState('');
  const fieldId = useId();

  if (itemId === null) return null;

  async function submit(event: FormEvent) {
    event.preventDefault();
    if (name.trim() === '') return;
    if (await create(name)) setName('');
  }

  return (
    <AppDialog title={t('save_title')} subject={itemTitle} busy={isCreating} onClose={close}>
      {isLoading && <AppSpinner label={t('common_loading')} />}
      {error !== null && <AppMessage tone="error" text={renderMessage(t, error)} />}
      {!isLoading && groups.length === 0 && <p className={styles.empty}>{t('save_empty')}</p>}

      {groups.length > 0 && (
        <ul className={styles.list}>
          {groups.map((group) => (
            <li key={group.id}>
              <label className={styles.option}>
                <input
                  type="checkbox"
                  className={styles.check}
                  checked={selected.includes(group.id)}
                  disabled={pending.includes(group.id)}
                  onChange={() => void toggle(group.id)}
                />
                <span className={styles.label}>{group.label}</span>
                <span className={styles.count}>{t('save_count', { count: group.count })}</span>
              </label>
            </li>
          ))}
        </ul>
      )}

      <form className={styles.create} onSubmit={(event) => void submit(event)}>
        <label className={fields.label} htmlFor={fieldId}>
          {t('save_new_label')}
        </label>
        <div className={styles.createRow}>
          <input
            id={fieldId}
            className={fields.field}
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder={t('save_new_placeholder')}
            maxLength={80}
            autoComplete="off"
            enterKeyHint="done"
          />
          <AppButton
            type="submit"
            variant="quiet"
            label={isCreating ? t('save_new_working') : t('save_new_submit')}
            disabled={isCreating || name.trim() === ''}
          />
        </div>
      </form>

      <div className={fields.actions}>
        <AppButton label={t('save_done')} onClick={close} disabled={isCreating} />
      </div>
    </AppDialog>
  );
}
