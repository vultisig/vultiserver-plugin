import Star from "@/assets/Star.svg?react";
import "./StarContainer.css";
import Button from "@/modules/core/components/ui/button/Button";
import { useState } from "react";

const STAR_MAX_COUNT = 5;

type StarContainerProps = {
  initialRating: number;
  disableChange?: boolean;
  onChange?: (change: number) => void;
};

const StarContainer = ({
  initialRating,
  disableChange,
  onChange,
}: StarContainerProps) => {
  const [rating, setRating] = useState<number>(initialRating);

  const handleChange = (change: number) => {
    if (!disableChange && onChange) {
      setRating(change);
      onChange(change);
    }
  };

  return (
    <div className="star-container">
      {[...Array(STAR_MAX_COUNT)].map((_, num) => (
        <Button
          key={num}
          size="mini"
          type="button"
          style={{ padding: "0px", cursor: disableChange ? "auto" : "pointer" }}
          styleType="tertiary"
          onClick={() => handleChange(num + 1)}
        >
          <Star
            data-testid={`rating-stars-${num}`}
            className={`star ${num + 1 <= rating ? "filled" : ""}`}
            width="24px"
            height="24px"
          />
        </Button>
      ))}
    </div>
  );
};

export default StarContainer;
