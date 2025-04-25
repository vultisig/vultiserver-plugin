import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import Review from "@/modules/review/components/review/Review";

describe("Review", () => {
  it("should render comment in the UI", () => {
    render(
      <Review
        id="1"
        date="2025-02-26T14:25:35Z"
        comment="comment text"
        rating={5}
      />
    );

    expect(screen.queryByText(`"comment text"`)).toBeInTheDocument();
  });

  it("should render ISO date in correct format ", () => {
    render(
      <Review
        id="1"
        date="2025-02-26T14:25:35Z"
        comment="comment text"
        rating={5}
      />
    );

    expect(screen.getByText("26.02.2025")).toBeInTheDocument();
  });
});
