import { useId } from 'react';
import { THEME_PREFERENCES, type ThemePreference } from '@domain/appearance/entities/theme';
import type { TranslationKey } from '@shared/i18n/keys';
import { useTranslate } from '@shared/i18n/use-translate';
import styles from './theme-picker.module.css';

const LABELS: Record<ThemePreference, TranslationKey> = {
  system: 'settings_theme_system',
  light: 'settings_theme_light',
  dark: 'settings_theme_dark',
};

type ThemePickerProps = {
  value: ThemePreference;
  onChange: (theme: ThemePreference) => void;
};

// Tre scelte che si escludono: radio veri, disegnati come un selettore a segmenti. La
// tastiera, l'annuncio e lo stato li dà l'HTML.
export function ThemePicker({ value, onChange }: ThemePickerProps) {
  const t = useTranslate();
  const name = useId();
  const hintId = useId();
  return (
    <fieldset className={styles.picker} aria-describedby={hintId}>
      <legend className={styles.legend}>{t('settings_theme')}</legend>
      <div className={styles.options}>
        {THEME_PREFERENCES.map((theme) => (
          <label key={theme} className={theme === value ? styles.optionActive : styles.option}>
            <input
              type="radio"
              name={name}
              className={styles.radio}
              value={theme}
              checked={theme === value}
              onChange={() => onChange(theme)}
            />
            {t(LABELS[theme])}
          </label>
        ))}
      </div>
      <p id={hintId} className={styles.hint}>
        {t('settings_theme_hint')}
      </p>
    </fieldset>
  );
}
