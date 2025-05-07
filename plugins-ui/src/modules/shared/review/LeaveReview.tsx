import Star from "@/assets/Star.svg?react";
import "./Review.css"; // Import CSS file for styling
import { useState } from "react";
import Button from "@/modules/core/components/ui/button/Button";
import DOMPurify from "dompurify";

const STAR_MAX_COUNT = 5;

const LeaveReview = () => {
  const [rating, setRating] = useState<number>(0);
  const [input, setInput] = useState("");

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const sanitizedText = DOMPurify.sanitize(e.target.value);
    setInput(sanitizedText);
  };

  return (
    <section className="leave-review" data-testid="leave-review-wrapper">
      <section className="review-score">
        <label className="label">Leave a review</label>
        <div className="star-container">
          {[...Array(STAR_MAX_COUNT)].map((_, num) => (
            <Button
              key={num}
              size="mini"
              type="button"
              style={{ padding: "0px" }}
              styleType="tertiary"
              onClick={() => setRating(num + 1)}
            >
              <Star
                className={`star ${num + 1 <= rating ? "filled" : ""}`}
                width="24px"
                height="24px"
              />
            </Button>
          ))}
        </div>
      </section>
      <textarea
        cols={80}
        className="review-textarea"
        placeholder="Install the plugin to leave a review"
        value={input}
        onChange={handleChange}
      ></textarea>

      <Button
        className="review-button"
        size="medium"
        type="button"
        styleType="primary"
        onClick={() => console.log("TODO make request")}
      >
        Leave a review
      </Button>
    </section>
  );
};

export default LeaveReview;
