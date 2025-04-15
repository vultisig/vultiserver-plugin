import Star from "@/assets/Star.svg?react";

const STAR_MAX_COUNT = 5;

type ReviewProps = {
  id: number;
  wallet: string;
  date: string;
  rating: number;
  comment: string;
};

const Review = ({ id, wallet, date, rating, comment }: ReviewProps) => {
  return (
    <>
      <div className="review-info-header" key={id}>
        <div className="review-icon"></div>
        <div>{wallet}</div>
        <div>{date}</div>
        <div className="star-container">
          {[...Array(STAR_MAX_COUNT)].map((_, num) => (
            <Star
              key={num}
              className={`star ${num + 1 <= rating ? "filled" : ""}`}
              width="24px"
              height="24px"
            />
          ))}
        </div>
      </div>
      <div>{`"${comment}"`}</div>
    </>
  );
};

export default Review;
