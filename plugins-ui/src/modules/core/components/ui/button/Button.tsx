import { HTMLProps } from "react";
import "./Button.css";

interface IButtonProps
  extends Omit<HTMLProps<HTMLButtonElement>, "size" | "type"> {
  type: "button" | "submit";
  styleType: "primary" | "secondary" | "tertiary";
  size: "mini" | "small" | "medium";
  ariaLabel?: string;
}

const Button = (props: IButtonProps) => {
  const {
    type,
    styleType,
    size,
    children,
    className,
    style,
    ariaLabel,
    onClick,
    ...rest
  } = props;
  return (
    <button
      type={type}
      onClick={onClick}
      className={`button ${styleType} ${size} ${className}`}
      style={style}
      aria-label={ariaLabel}
      {...rest}
    >
      {children}
    </button>
  );
};

export default Button;
