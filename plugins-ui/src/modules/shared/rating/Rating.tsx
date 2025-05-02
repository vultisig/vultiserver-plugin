import Star from "@/assets/Star.svg?react";
import "./Rating.css";
import { useReviews } from "@/modules/review/context/ReviewProvider";

const formatNumber = (num: number): string | number => {
  return Number(num.toFixed(2).replace(/\.?0+$/, ""));
};

const MAX_BAR_LENGTH = 336;

const Rating = () => {
  const { pluginRatings } = useReviews();
  const ratingsMap = new Map(
    pluginRatings?.map((ratingObj) => [ratingObj.rating, ratingObj])
  );

  const rating_bars = [5, 4, 3, 2, 1];

  const { ratingsCount, ratingSum, maxRatingCount } = pluginRatings.reduce(
    (acc, rating) => {
      acc.ratingsCount += rating.count;
      acc.ratingSum += rating.rating * rating.count;
      if (rating.count > acc.maxRatingCount) {
        acc.maxRatingCount = rating.count;
      }
      return acc;
    },
    { ratingsCount: 0, ratingSum: 0, maxRatingCount: 0 }
  );

  const averageRating =
    ratingSum > 0 && ratingsCount > 0
      ? formatNumber(ratingSum / ratingsCount)
      : 0;

  return (
    <section className="rating">
      <div className="rating-chart">
        {rating_bars &&
          rating_bars.map((r) => (
            <div key={r} className="rating-bar">
              <div className="stars">{r}</div>
              <div className="bar">
                <div className="bar-track"></div>
                <div
                  className="bar-handle"
                  style={{
                    width: `${(ratingsMap.get(r)?.count || 0) * (MAX_BAR_LENGTH / maxRatingCount)}px`,
                  }}
                ></div>
              </div>
            </div>
          ))}
      </div>
      <div className="rating-summary">
        <div className="rating-average">
          {averageRating} &nbsp;
          <Star className={"star filled"} width="20px" height="20px" />
        </div>
        <div className="rating-sum">
          {ratingsCount}&nbsp; <span>Reviews</span>
        </div>
      </div>
    </section>
  );
};

export default Rating;
