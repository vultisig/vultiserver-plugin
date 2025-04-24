import LeaveReview from "@/modules/review/components/leave-review/LeaveReview";
import ReviewHistory from "@/modules/review/components/review-history/ReviewHistory";
import "./Reviews.css";
import { ReviewProvider } from "../../context/ReviewProvider";
import { Plugin } from "@/modules/plugin/models/plugin";
import Rating from "@/modules/shared/rating/Rating";

type ReviewProps = {
  plugin: Plugin;
};
const Reviews = ({ plugin }: ReviewProps) => {
  return (
    <ReviewProvider pluginId={plugin.id} ratings={plugin.ratings}>
      <section>
        <h3 className="review-rating-header">Reviews and Ratings</h3>
        <div className="review-rating">
          <LeaveReview />
          <Rating />
        </div>
      </section>

      <section className="reviews">
        <ReviewHistory />
      </section>
    </ReviewProvider>
  );
};

export default Reviews;
