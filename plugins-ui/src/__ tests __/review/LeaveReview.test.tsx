import { describe, expect, it, Mock, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import LeaveReview from "@/modules/shared/review/LeaveReview";
import { useReviews } from "@/modules/review/context/ReviewProvider";

vi.mock("@/modules/review/context/ReviewProvider", () => ({
  useReviews: vi.fn(),
}));

describe("LeaveReview", () => {
  it("should render Leave a review button as disabled when rating or input is missing", async () => {
    (useReviews as Mock).mockReturnValue({
      reviewsMap: {
        reviews: [
          { id: "1", created_at: "2024-04-01", rating: 5, comment: "Great!" },
          {
            id: "2",
            created_at: "2024-03-29",
            rating: 4,
            comment: "Good service",
          },
        ],
      },
      page: 1,
      setPage: vi.fn(),
      totalPages: 1,
    });

    render(<LeaveReview />);

    const button = screen.getByRole("button", { name: /Leave a review/i });

    expect(button).toHaveClass("disabled");
  });

  it("should render Leave a review button not as disabled when rating & input are present", async () => {
    const mockAddReview = vi.fn().mockResolvedValue(true);

    (useReviews as Mock).mockReturnValue({
      reviewsMap: {
        reviews: [
          { id: "1", created_at: "2024-04-01", rating: 5, comment: "Great!" },
          {
            id: "2",
            created_at: "2024-03-29",
            rating: 4,
            comment: "Good service",
          },
        ],
      },
      page: 1,
      setPage: vi.fn(),
      totalPages: 1,
      pluginId: "mock-plugin-id",
      addReview: mockAddReview,
    });

    render(<LeaveReview />);

    const input = screen.getByRole("textbox");
    const ratingStars = screen.getByTestId("rating-stars-2");

    fireEvent.change(input, {
      target: { value: "This is a great review!" },
    });
    fireEvent.click(ratingStars);

    const button = screen.getByRole("button", { name: /Leave a review/i });

    fireEvent.click(button);

    expect(button).not.toHaveClass("disabled");
  });
});
