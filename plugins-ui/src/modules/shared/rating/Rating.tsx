import Star from "@/assets/Star.svg?react";
import "./Rating.css";

const Rating = () => {
  const ratings = [5, 4, 3, 2, 1];

  return (
    <section className="rating">
      <div className="rating-chart">
        {ratings &&
          ratings.map((r) => (
            <div className="rating-bar" key={r}>
              <div className="count">{r}</div>
              <div className="bar" style={{ width: `${r * 63}px` }}></div>
            </div>
          ))}
      </div>
      {4.5}
      <Star className={"star filled"} width="20px" height="20px" />
    </section>
  );
};

export default Rating;
