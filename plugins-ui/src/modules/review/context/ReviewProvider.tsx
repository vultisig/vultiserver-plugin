import {
  createContext,
  ReactNode,
  useContext,
  useEffect,
  useState,
} from "react";
import Toast from "@/modules/core/components/ui/toast/Toast";
import MarketplaceService from "@/modules/marketplace/services/marketplaceService";
import {
  CreateReview,
  ReviewMap,
} from "@/modules/marketplace/models/marketplace";
import { PluginRatings } from "@/modules/plugin/models/plugin";

const ITEMS_PER_PAGE = 5;

export interface ReviewContextType {
  pluginId: string;
  pluginRatings: PluginRatings[];
  reviewsMap: ReviewMap | undefined;
  page: number;
  setPage: (page: number) => void;
  totalPages: number;
  addReview: (pluginId: string, review: CreateReview) => Promise<boolean>;
}

export const ReviewContext = createContext<ReviewContextType | undefined>(
  undefined
);

type ReviewProviderProps = {
  pluginId: string;
  ratings: PluginRatings[];
  children: ReactNode;
};

export const ReviewProvider = ({
  pluginId,
  ratings,
  children,
}: ReviewProviderProps) => {
  const [reviewsMap, setReviewsMap] = useState<ReviewMap>();
  const [page, setPage] = useState(1);
  const [pluginRatings, setPluginRatings] = useState(ratings);
  const [totalPages, setTotalPages] = useState(1);

  const [toast, setToast] = useState<{
    message: string;
    error?: string;
    type: "success" | "error";
  } | null>(null);

  useEffect(() => {
    if (!pluginId) return;
    const skip = (page - 1) * ITEMS_PER_PAGE;

    const fetchReviews = async (): Promise<void> => {
      try {
        const fetchedReviews = await MarketplaceService.getReviews(
          pluginId,
          skip,
          ITEMS_PER_PAGE
        );

        setTotalPages(Math.ceil(fetchedReviews.total_count / ITEMS_PER_PAGE));

        setReviewsMap(fetchedReviews);
      } catch (error: any) {
        console.error("Failed to get reviews:", error.message);
        setToast({
          message: error.message || "Failed to get reviews",
          error: error.error,
          type: "error",
        });
      }
    };

    fetchReviews();
  }, [page]); // Refetch when page changes

  const addReview = async (
    pluginId: string,
    review: CreateReview
  ): Promise<boolean> => {
    try {
      const newReview = await MarketplaceService.createReview(pluginId, review);
      setPluginRatings(newReview.ratings);

      setReviewsMap((prev) => {
        if (!prev) return { reviews: [newReview], total_count: 1 }; // Handle initial case

        const newTotalCount = prev.total_count + 1;

        if (newTotalCount / ITEMS_PER_PAGE > 1) {
          setTotalPages(Math.ceil(newTotalCount / ITEMS_PER_PAGE));
        }

        return {
          ...prev,
          reviews: [
            newReview,
            ...(prev?.reviews?.slice(0, ITEMS_PER_PAGE - 1) || []),
          ], // Add new review
          total_count: newTotalCount,
        };
      });
      setToast({ message: "Review created successfully!", type: "success" });

      return Promise.resolve(true);
    } catch (error: any) {
      console.error("Failed to create review:", error.message);
      setToast({
        message: error.message || "Failed to create review",
        error: error.error,
        type: "error",
      });

      return Promise.resolve(false);
    }
  };

  return (
    <ReviewContext.Provider
      value={{
        pluginId,
        pluginRatings,
        reviewsMap,
        page,
        setPage,
        totalPages,
        addReview,
      }}
    >
      {children}
      {toast && (
        <Toast
          title={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}
    </ReviewContext.Provider>
  );
};

export const useReviews = (): ReviewContextType => {
  const context = useContext(ReviewContext);
  if (!context) {
    throw new Error("useReviews must be used within a ReviewProvider");
  }
  return context;
};
