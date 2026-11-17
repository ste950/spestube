import styles from './app-button.module.css';

type AppButtonProps = {
  label: string;
  onClick?: () => void;
  variant?: 'primary' | 'quiet';
  disabled?: boolean;
  /** `submit` solo dentro un form: il default del browser è `submit`, ed è quasi sempre sbagliato. */
  type?: 'button' | 'submit';
};

export function AppButton({
  label,
  onClick,
  variant = 'primary',
  disabled = false,
  type = 'button',
}: AppButtonProps) {
  return (
    <button
      type={type}
      className={variant === 'primary' ? styles.primary : styles.quiet}
      onClick={onClick}
      disabled={disabled}
    >
      {label}
    </button>
  );
}
