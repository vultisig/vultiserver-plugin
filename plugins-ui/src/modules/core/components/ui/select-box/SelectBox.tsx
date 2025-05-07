import { CSSProperties, useState } from "react";
import ChevronDown from "@/assets/ChevronDown.svg?react";
import "./SelectBox.css";

type SelectBoxProps = {
  options: string[];
  onSelectChange: (option: string) => void;
  value?: string;
  label?: string;
  style?: CSSProperties;
};

const SelectBox = ({
  label,
  options,
  value,
  onSelectChange,
  style,
}: SelectBoxProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedOption, setSelectedOption] = useState(value || "");

  const handleSelect = (option: string) => {
    onSelectChange(option);
    setSelectedOption(option);
    setIsOpen(false);
  };
  return (
    <div style={style} className="custom-dropdown">
      <div
        className="dropdown-toggle"
        onClick={() => setIsOpen(!isOpen)}
        data-testid="select-box-trigger"
      >
        {label}
        <div className="toggle-choice" data-testid="select-box-selected">
          {selectedOption}
          <ChevronDown width="20px" height="20px" />
        </div>
      </div>
      {isOpen && (
        <ul className={`dropdown-menu open`}>
          {options.map((option) => (
            <li
              key={option}
              onClick={() => handleSelect(option)}
              data-testid="select-box-option"
            >
              {option}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default SelectBox;
