import styles from './app-message.module.css';

type AppMessageProps = {
  text: string;
  tone?: 'neutral' | 'error' | 'success';
};

// Un errore si annuncia subito (`alert`), il resto quando il lettore di schermo ha finito
// (`status`): un avviso che interrompe per dire «nessun video» è rumore.
export function AppMessage({ text, tone = 'neutral' }: AppMessageProps) {
  return (
    <p className={styles[tone]} role={tone === 'error' ? 'alert' : 'status'}>
      {text}
    </p>
  );
}
