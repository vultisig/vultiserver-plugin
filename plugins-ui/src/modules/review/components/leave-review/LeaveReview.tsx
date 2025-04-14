import "./LeaveReview.css";
import { useState } from "react";
import Button from "@/modules/core/components/ui/button/Button";
import { CreateReview } from "@/modules/marketplace/models/marketplace";
import { useReviews } from "../../context/ReviewProvider";
import StarContainer from "@/modules/shared/star-container/StartContainer";

const LeaveReview = () => {
  const { pluginId, addReview } = useReviews();

  const [input, setInput] = useState("");
  const [rating, setRating] = useState(0);

  const submitReview = () => {
    if (rating && input) {
      const review: CreateReview = {
        address: "TODO", // todo remove this when we have the participant from installation
        comment: input,
        rating: rating,
      };

      addReview(pluginId, review).then((reviewAdded) => {
        if (reviewAdded) {
          setInput("");
          setRating(0);
        }
      });
    }
  };

  return (
    <section className="leave-review" data-testid="leave-review-wrapper">
      <section className="review-score">
        <label className="label">Leave a review</label>

        <StarContainer
          key={rating}
          initialRating={rating}
          onChange={setRating}
        />
      </section>
      <textarea
        cols={78}
        className="review-textarea"
        placeholder="Install the plugin to leave a review"
        value={input}
        onChange={(e) => setInput(e.target.value)}
      ></textarea>

      <Button
        className={`review-button ${!rating || !input ? "disabled" : ""}`}
        size="medium"
        type="button"
        styleType="primary"
        onClick={submitReview}
      >
        Leave a review
      </Button>
    </section>
  );
};

export default LeaveReview;
