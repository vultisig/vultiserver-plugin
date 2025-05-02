import { useEffect, useState } from "react";
import Review from "../review/Review";
import "./ReviewHistory.css";
import { Review as ReviewType } from "@/modules/marketplace/models/marketplace";
import Pagination from "@/modules/core/components/ui/pagination/Pagination";
import { useReviews } from "@/modules/review/context/ReviewProvider";

const ReviewHistory = () => {
  const { reviewsMap, page, setPage, totalPages } = useReviews();
  const [reviewHistory, setReviewHistory] = useState<ReviewType[] | undefined>(
    []
  );

  useEffect(() => {
    setReviewHistory(reviewsMap?.reviews ? [...reviewsMap.reviews] : []);
  }, [reviewsMap]);

  return (
    <>
      <section className="review-history">
        {reviewHistory &&
          reviewHistory.length > 0 &&
          reviewHistory.map((review) => (
            <Review
              key={review.id}
              id={review.id}
              date={review.created_at}
              rating={review.rating}
              comment={review.comment}
            />
          ))}
      </section>
      {totalPages > 1 && (
        <Pagination
          currentPage={page}
          totalPages={totalPages}
          onPageChange={setPage}
        />
      )}
    </>
  );
};

export default ReviewHistory;
