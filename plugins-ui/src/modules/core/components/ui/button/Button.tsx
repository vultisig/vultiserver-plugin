import { ReactNode } from "react";
import "./Button.css";

type ButtonProps = {
  type: "button" | "submit";
  styleType: "primary" | "secondary" | "tertiary";
  size: "mini" | "small" | "medium";
  children: ReactNode;
  className?: string;
  style?: {};
  ariaLabel?: string;
  onClick?: () => any;
  disabled?: boolean;
};

const Button = ({
  type,
  styleType,
  size,
  children,
  className = "",
  style,
  ariaLabel,
  onClick,
  disabled = false,
}: ButtonProps) => {
  return (
    <button
      type={type}
      onClick={onClick}
      className={`button ${styleType} ${size} ${className}`}
      style={style}
      aria-label={ariaLabel}
      disabled={disabled}
    >
      {children}
    </button>
  );
};

export default Button;
