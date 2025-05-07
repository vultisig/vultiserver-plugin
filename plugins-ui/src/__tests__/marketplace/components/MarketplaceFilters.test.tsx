import MarketplaceFilters from "@/modules/marketplace/components/marketplace-filters/MarketplaceFilters";
import { render, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";

const FILTERS = {
  term: "",
  categoryId: "1",
  sortBy: "",
  sortOrder: "asc",
};

const mockOnFiltersChange = vi.fn();

const mockOnViewChange = vi.fn();
describe("MarketplaceFilters", async () => {
  it("should call onChange with 'grid'", async () => {
    const { findByTestId } = render(
      <MarketplaceFilters
        viewFilter="list"
        categories={[]}
        filters={FILTERS}
        onFiltersChange={mockOnFiltersChange}
        onViewChange={mockOnViewChange}
      />
    );
    const gridButton = await findByTestId("marketplace-filters-grid");
    await userEvent.click(gridButton);
    await waitFor(() => {
      expect(mockOnViewChange).toBeCalledWith("grid");
    });
  });
  it("should call onChange with 'list'", async () => {
    const { findByTestId } = render(
      <MarketplaceFilters
        viewFilter="grid"
        categories={[]}
        filters={FILTERS}
        onFiltersChange={mockOnFiltersChange}
        onViewChange={mockOnViewChange}
      />
    );
    const listButton = await findByTestId("marketplace-filters-list");
    await userEvent.click(listButton);
    await waitFor(() => {
      expect(mockOnViewChange).toBeCalledWith("list");
    });
  });
});
