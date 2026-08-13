import { memo } from 'react';
import styles from './Styles.module.scss';
import cn from 'classnames';

interface InputProps {
  placeholder?: string;
  value?: string | number; 
  onChange?: (value: string | number) => void; 
  label?: string;
  error?: string;  
  type?: 'text' | 'number' | 'search';
  disabled?: boolean;
  className?: string;
  name?: string;
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
  onClick?: (e: React.MouseEvent<HTMLInputElement>) => void;
  inputMode?: React.HTMLAttributes<HTMLInputElement>["inputMode"];
}

function InputComponent({ 
  placeholder, 
  value = '', 
  onChange = () => {}, 
  label, 
  error, 
  type = 'text', 
  disabled, 
  className,
  onKeyDown,
  onClick,
  inputMode,
  ...props
}: InputProps) {
  return (
    <div className={cn(styles.wrapper, className)} onClick={(e) => e.stopPropagation()}>
      {label && <label className={styles.wrapper__label}>{label}</label>}
      <input
        type={type}
        className={cn(styles.wrapper__input, { [styles.error]: error })}
        placeholder={placeholder}
        value={value}
        onChange={(e) => {
          const val = e.target.value;
          if (type === 'number') {
            onChange(val === '' ? '' : Number(val));
          } else {
            onChange(val);
          }
        }}
        onKeyDown={onKeyDown}
        onClick={onClick}
        inputMode={inputMode}
        disabled={disabled}
        {...props}
      />
      {error && <p className={styles.wrapper__error}>{error}</p>}
    </div>
  );
}

export const Input = memo(InputComponent);