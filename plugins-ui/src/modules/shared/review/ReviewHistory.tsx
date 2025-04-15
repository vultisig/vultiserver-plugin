import Review from "./Review";
import "./Review.css";

// todo remove hardcoding once endpoints are ready
const ReviewHistory = () => {
  const reviewHistory = [1, 2, 3, 4, 5];
  return (
    <section className="review-history">
      {reviewHistory &&
        reviewHistory.map((r) => (
          <Review
            key={r}
            id={r}
            wallet="0x12324...34423"
            date="05.01.2025"
            rating={3}
            comment="There is no one who loves pain itself, who seeks after it and wants to
        have it, simply because it is pain..."
          />
        ))}
    </section>
  );
};

export default ReviewHistory;
