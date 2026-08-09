import { memo } from 'react';
import styles from './Styles.module.scss';
import cn from 'classnames';

interface ToggleSwitchProps {
  checked: boolean;
  disabled?: boolean;
  onChange: (checked: boolean) => void;
}

function ToggleSwitchComponent({ checked, disabled, onChange }: ToggleSwitchProps) {
  return (
    <div className={styles.toggleWrapper}>
      <div
        className={cn(
          styles.toggleWrapper__toggle,
          checked && styles.toggleWrapper__toggle__on,
          disabled && styles.toggleWrapper__toggle__disabled
        )}
        onClick={() => !disabled && onChange(!checked)}
      >
        <div className={cn(styles.toggleWrapper__knob, checked && styles.toggleWrapper__knob__on)} />
      </div>
    </div>
  );
}

export const ToggleSwitch = memo(ToggleSwitchComponent);