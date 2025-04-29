import {
  fireEvent,
  render,
  renderHook,
  screen,
  waitFor,
} from "@testing-library/react";
import { describe, it, expect, vi, Mock } from "vitest";
import MarketplaceService from "@/modules/marketplace/services/marketplaceService";
import { ReviewMap } from "@/modules/marketplace/models/marketplace";
import {
  ReviewProvider,
  useReviews,
} from "@/modules/review/context/ReviewProvider";
const ratings = [
  { rating: 1, count: 0 },
  { rating: 2, count: 0 },
  { rating: 3, count: 0 },
  { rating: 4, count: 0 },
  { rating: 5, count: 0 },
];

const mockReviewMap: ReviewMap = {
  reviews: [
    {
      id: "1",
      address: "TODO",
      comment: "comment 1",
      rating: 5,
      created_at: "date",
      plugin_id: "123",
      ratings: ratings,
    },
    {
      id: "2",
      address: "TODO",
      comment: "comment 2",
      rating: 3,
      created_at: "date",
      plugin_id: "123",
      ratings: ratings,
    },
  ],
  total_count: 2,
};

vi.mock("@/modules/marketplace/services/marketplaceService", () => ({
  default: {
    getReviews: vi.fn(),
    createReview: vi.fn(),
  },
}));

const hoisted = vi.hoisted(() => ({
  mockEventBus: {
    publish: vi.fn(),
  },
}));

vi.mock("@/utils/eventBus", () => ({
  publish: hoisted.mockEventBus.publish,
}));

const TestComponent = ({ pluginId }: { pluginId: string }) => {
  const { reviewsMap, addReview, totalPages } = useReviews();

  return (
    <div>
      <ul>
        {reviewsMap?.reviews.map((review) => (
          <li key={review.id}>{review.id}</li>
        ))}
      </ul>

      <button
        onClick={() =>
          addReview(pluginId, {
            address: "TODO",
            comment: "some review comments",
            rating: 5,
          })
        }
      >
        Add Review
      </button>
      <div>Total pages: {totalPages}</div>
    </div>
  );
};

const renderWithProvider = () => {
  return render(
    <ReviewProvider pluginId="123" ratings={ratings}>
      <TestComponent pluginId="123" />
    </ReviewProvider>
  );
};

describe("ReviewProvider", () => {
  describe("useReviews Hook", () => {
    it("should throw an error if used outside of ReviewProvider", () => {
      expect(() => renderHook(() => useReviews())).toThrow(
        "useReviews must be used within a ReviewProvider"
      );
    });
  });

  describe("getReviews", () => {
    it("should fetch & store review in context", async () => {
      (MarketplaceService.getReviews as Mock).mockResolvedValue(mockReviewMap);
      renderWithProvider();

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
      });
    });

    it("should handle API failure and set toast error when getPolicies request fails", async () => {
      const mockError = new Error("API Error");

      (MarketplaceService.getReviews as Mock).mockRejectedValue(mockError);

      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      renderWithProvider();

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalledWith(
          "Failed to get reviews:",
          "API Error"
        );
        expect(hoisted.mockEventBus.publish).toHaveBeenCalledWith("onToast", {
          message: "API Error",
          type: "error",
        });
      });
    });
  });

  describe("createReview", () => {
    it("should add review in context", async () => {
      (MarketplaceService.getReviews as Mock).mockResolvedValue(mockReviewMap);

      (MarketplaceService.createReview as Mock).mockResolvedValue({
        id: "3",
        address: "TODO",
        comment: "comment 1",
        rating: 5,
        created_at: "date",
        plugin_id: "123",
      });

      renderWithProvider();

      const newReviewButton = screen.getByRole("button", {
        name: "Add Review",
      });

      fireEvent.click(newReviewButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(screen.getByText("3")).toBeInTheDocument();
        expect(screen.getByText("Total pages: 1")).toBeInTheDocument();
        expect(hoisted.mockEventBus.publish).toBeCalledWith("onToast", {
          message: "Review created successfully!",
          type: "success",
        });
      });
    });

    it("should update pagination when review get more than 5 on a page", async () => {
      const reviewsMap: ReviewMap = {
        reviews: [
          {
            id: "1",
            address: "TODO",
            comment: "comment 1",
            rating: 5,
            created_at: "date",
            plugin_id: "123",
            ratings: ratings,
          },
          {
            id: "2",
            address: "TODO",
            comment: "comment 2",
            rating: 3,
            created_at: "date",
            plugin_id: "123",
            ratings: ratings,
          },
          {
            id: "3",
            address: "TODO",
            comment: "comment 1",
            rating: 5,
            created_at: "date",
            plugin_id: "123",
            ratings: ratings,
          },
          {
            id: "4",
            address: "TODO",
            comment: "comment 1",
            rating: 5,
            created_at: "date",
            plugin_id: "123",
            ratings: ratings,
          },
          {
            id: "5",
            address: "TODO",
            comment: "comment 1",
            rating: 5,
            created_at: "date",
            plugin_id: "123",
            ratings: ratings,
          },
        ],
        total_count: 5,
      };
      (MarketplaceService.getReviews as Mock).mockResolvedValue(reviewsMap);

      (MarketplaceService.createReview as Mock).mockResolvedValue({
        id: "6",
        address: "TODO",
        comment: "comment 1",
        rating: 5,
        created_at: "date",
        plugin_id: "123",
      });

      renderWithProvider();

      const newReviewButton = screen.getByRole("button", {
        name: "Add Review",
      });

      fireEvent.click(newReviewButton);

      await waitFor(() => {
        expect(screen.getByText("Total pages: 2")).toBeInTheDocument();
      });
    });

    it("should handle API failure and set toast error when createReview request fails", async () => {
      (MarketplaceService.getReviews as Mock).mockResolvedValue(mockReviewMap);
      (MarketplaceService.createReview as Mock).mockRejectedValue("API Error");

      renderWithProvider();

      const newReviewButton = screen.getByRole("button", {
        name: "Add Review",
      });

      fireEvent.click(newReviewButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(screen.queryByText("3")).not.toBeInTheDocument();
      });
    });
  });
});
