import type { ReactNode } from 'react';
import styles from './app-visually-hidden.module.css';

type AppVisuallyHiddenProps = {
  children: ReactNode;
  /** `h1` per il titolo di una schermata che a video non serve, ma a chi ascolta sì. */
  as?: 'span' | 'h1';
};

// Testo che non si vede ma si legge: per ciò che la vista capisce dal contesto (un numero
// accanto a un'icona, la sezione accesa nella barra) e un lettore di schermo no.
export function AppVisuallyHidden({ children, as: Element = 'span' }: AppVisuallyHiddenProps) {
  return <Element className={styles.hidden}>{children}</Element>;
}
