import { memo } from "react";
import styles from "./Styles.module.scss";
import cn from "classnames";

interface ButtonProps {
  size?: "nav" | "main" | "long" | "min";
  color?: "primary" | "transparent";
  children?: React.ReactNode;
  onClick?: () => void;
  type?: "button" | "submit" | "reset";
  className?: string;
  disabled?: boolean;
}

function ButtonComponent({ size, color, children, onClick, type = "button", className = "", disabled, }: ButtonProps) {
  return (
    <button
      onClick={onClick}
      type={type}
      className={cn(styles.button, styles[`button__${color}`], styles[`button__${size}`], className,)}
      disabled={disabled}
    >
      {children}
    </button>
  );
}

export const Button = memo(ButtonComponent);