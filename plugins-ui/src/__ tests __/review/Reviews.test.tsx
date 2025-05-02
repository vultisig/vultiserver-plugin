import { describe, expect, it, Mock, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { Plugin } from "@/modules/plugin/models/plugin";
import { useReviews } from "@/modules/review/context/ReviewProvider";
import Reviews from "@/modules/review/components/reviews/Reviews";

const plugin: Plugin = {
  id: "mock-plugin-id",
  type: "type",
  title: "title",
  description: "description",
  metadata: {},
  server_endpoint: "endpoint",
  pricing_id: "pricingId",
  tags: [],
  category_id: "categoryId",
  ratings: [
    { rating: 1, count: 0 },
    { rating: 2, count: 0 },
    { rating: 3, count: 0 },
    { rating: 4, count: 0 },
    { rating: 5, count: 0 },
  ],
};

vi.mock("@/modules/review/context/ReviewProvider", async (importOriginal) => {
  const actual = (await importOriginal()) as {};
  return {
    ...actual,
    useReviews: vi.fn(),
  };
});

describe("Reviews", () => {
  it("should update reviews when leaving a review", async () => {
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
      pluginRatings: plugin.ratings,
    });

    const { rerender } = render(<Reviews plugin={plugin} />);

    await waitFor(() => {
      expect(screen.getByText(`"Great!"`)).toBeInTheDocument();
      expect(screen.getByText(`"Good service"`)).toBeInTheDocument();
      expect(
        screen.queryByText(`"This is a great review!"`)
      ).not.toBeInTheDocument();
    });

    const input = screen.getByRole("textbox");
    const ratingStars = screen.getAllByTestId("rating-stars-2");

    fireEvent.change(input, {
      target: { value: "This is a great review!" },
    });

    ratingStars.forEach((star) => {
      fireEvent.click(star);
    });

    const button = screen.getByRole("button", { name: /Leave a review/i });

    fireEvent.click(button);

    expect(button).not.toHaveClass("disabled");

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
          {
            id: "3",
            created_at: "2024-03-29",
            rating: 5,
            comment: "This is a great review!",
          },
        ],
      },
      page: 1,
      setPage: vi.fn(),
      totalPages: 1,
      pluginId: "mock-plugin-id",
      addReview: mockAddReview,
      pluginRatings: plugin.ratings,
    });

    rerender(<Reviews plugin={plugin} />);

    await waitFor(() => {
      expect(screen.getByText(`"Great!"`)).toBeInTheDocument();
      expect(screen.getByText(`"Good service"`)).toBeInTheDocument();
      expect(screen.getByText(`"This is a great review!"`)).toBeInTheDocument();
    });
  });
});
