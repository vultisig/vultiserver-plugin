import Pagination from "@/modules/core/components/ui/pagination/Pagination";
import { render, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, vi, expect } from "vitest";

describe("Pagination", () => {
  it("should call onPageChange", async () => {
    const mockOnPageChange = vi.fn();
    const { findByTestId, findAllByTestId } = render(
      <Pagination
        currentPage={1}
        totalPages={10}
        onPageChange={mockOnPageChange}
      />
    );
    await findByTestId("pagination-prev");
    const nextButton = await findByTestId("pagination-next");
    const pages = await findAllByTestId("pagination-page");
    userEvent.click(nextButton);
    await waitFor(() => {
      expect(mockOnPageChange).toBeCalledWith(2);
      expect(pages.length).toEqual(7);
    });
  });
  it("should disable prev/next button based on current page", async () => {
    const mockOnPageChange = vi.fn();
    const { findByTestId, findAllByTestId, rerender } = render(
      <Pagination
        currentPage={1}
        totalPages={2}
        onPageChange={mockOnPageChange}
      />
    );
    const prevButton = (await findByTestId(
      "pagination-prev"
    )) as HTMLButtonElement;
    expect(prevButton.disabled).toBe(true);
    const nextButton = (await findByTestId(
      "pagination-next"
    )) as HTMLButtonElement;
    const pages = await findAllByTestId("pagination-page");
    userEvent.click(nextButton);
    await waitFor(() => {
      expect(mockOnPageChange).toBeCalledWith(2);
      expect(pages.length).toEqual(2);
      rerender(
        <Pagination
          currentPage={2}
          totalPages={2}
          onPageChange={mockOnPageChange}
        />
      );
      expect(nextButton.disabled).toBe(true);
    });
  });
});
