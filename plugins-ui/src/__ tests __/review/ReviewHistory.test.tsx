import { describe, expect, it, Mock, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import ReviewHistory from "@/modules/review/components/review-history/ReviewHistory";
import { useReviews } from "@/modules/review/context/ReviewProvider";

vi.mock("@/modules/review/context/ReviewProvider", () => ({
  useReviews: vi.fn(),
}));

describe("ReviewHistory", () => {
  it("should render reviews correctly", async () => {
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

    render(<ReviewHistory />);

    await waitFor(() => {
      expect(screen.getByText(`"Great!"`)).toBeInTheDocument();
      expect(screen.getByText(`"Good service"`)).toBeInTheDocument();
    });
  });

  it("should show pagination if totalPages > 1", async () => {
    (useReviews as Mock).mockReturnValue({
      reviewsMap: { reviews: [] },
      page: 1,
      setPage: vi.fn(),
      totalPages: 2,
    });

    render(<ReviewHistory />);

    await waitFor(() => {
      expect(screen.getByTestId("pagination-wrapper")).toBeInTheDocument();
    });
  });

  it("should not show pagination if totalPages <= 1", async () => {
    (useReviews as Mock).mockReturnValue({
      reviewsMap: { reviews: [] },
      page: 1,
      setPage: vi.fn(),
      totalPages: 1,
    });

    render(<ReviewHistory />);

    await waitFor(() => {
      expect(
        screen.queryByTestId("pagination-wrapper")
      ).not.toBeInTheDocument();
    });
  });
});
