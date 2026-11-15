import { useEffect, useId, useRef, type ReactNode } from 'react';
import styles from './app-dialog.module.css';

type AppDialogProps = {
  title: string;
  /** A cosa si riferisce la finestra, per esempio il titolo di un video. */
  subject?: string;
  /** Mentre è vero, Esc non chiude: un salvataggio in corso non si interrompe a metà. */
  busy?: boolean;
  onClose: () => void;
  children: ReactNode;
};

// La finestra modale dell'app. Si monta quando serve e si smonta quando ha finito: chi la
// usa non tiene un «aperta/chiusa», la rende o non la rende.
export function AppDialog({ title, subject, busy = false, onClose, children }: AppDialogProps) {
  const dialog = useRef<HTMLDialogElement>(null);
  const titleId = useId();

  // Un `<dialog>` vero, aperto come modale: il focus ci resta dentro, il resto della pagina
  // diventa inerte e Esc lo chiude, senza ricostruire niente di tutto questo a mano.
  useEffect(() => {
    const element = dialog.current;
    if (element === null || element.open) return;
    element.showModal();
    return () => element.close();
  }, []);

  return (
    <dialog
      ref={dialog}
      className={styles.dialog}
      aria-labelledby={titleId}
      // Esc chiede di chiudere: la decisione resta a chi ha aperto la finestra, che così sa
      // sempre se è aperta.
      onCancel={(event) => {
        event.preventDefault();
        if (!busy) onClose();
      }}
    >
      <div className={styles.body}>
        <h2 id={titleId} className={styles.title}>
          {title}
        </h2>
        {subject !== undefined && <p className={styles.subject}>{subject}</p>}
        {children}
      </div>
    </dialog>
  );
}
