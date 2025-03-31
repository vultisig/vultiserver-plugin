import Star from "@/assets/Star.svg?react";
import "./Review.css"; // Import CSS file for styling
import { useState } from "react";
import Button from "@/modules/core/components/ui/button/Button";

// todo make sure to sanitize input

const STAR_MAX_COUNT = 5;

const LeaveReview = () => {
  const [rating, setRating] = useState<number>(0);
  return (
    <>
      <section>
        Leave a review {rating}
        <div className="star-container">
          {[...Array(STAR_MAX_COUNT)].map((_, num) => (
            <Button
              key={num}
              size="mini"
              type="button"
              style={{ paddingLeft: "0px", paddingTop: "2rem" }}
              styleType="tertiary"
              onClick={() => setRating(num + 1)}
            >
              <Star
                className={`star ${num + 1 <= rating ? "filled" : ""}`}
                width="20px"
                height="20px"
              />
            </Button>
          ))}
        </div>
      </section>
      <textarea></textarea>
    </>
  );
};

export default LeaveReview;
