import StarContainer from "@/modules/shared/star-container/StartContainer";
import "./Review.css";

type ReviewProps = {
  id: string;
  date: string;
  rating: number;
  comment: string;
};

const formatDate = (isoString: string): string => {
  const date = new Date(isoString);
  const day = String(date.getDate()).padStart(2, "0");
  const month = String(date.getMonth() + 1).padStart(2, "0"); // Months are 0-based
  const year = date.getFullYear();

  return `${day}.${month}.${year}`;
};

const Review = ({ id, date, rating, comment }: ReviewProps) => {
  return (
    <div className="single-review">
      <div className="review-info-header" key={id}>
        <div className="review-icon"></div>
        <div className="review-date">{formatDate(date)}</div>
        <StarContainer initialRating={rating} disableChange={true} />
      </div>
      <div>{`"${comment}"`}</div>
    </div>
  );
};

export default Review;
