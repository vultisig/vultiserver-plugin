import Modal from "@/modules/core/components/ui/modal/Modal";
import { render, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

describe("Modal", () => {
  it("should call onClose", async () => {
    const mockOnClose = vi.fn();
    const { findByTestId } = render(
      <Modal isOpen={true} onClose={mockOnClose}>
        <div data-testid="content">
          <h1>Modal content</h1>
        </div>
      </Modal>
    );
    await findByTestId("modal-overlay");
    await findByTestId("modal-content");
    await findByTestId("content");
    const button = await findByTestId("modal-close-button");
    userEvent.click(button);
    await waitFor(() => {
      expect(mockOnClose).toBeCalled();
    });
  });
  it("should not be visible", async () => {
    const mockOnClose = vi.fn();
    const { queryByTestId } = render(
      <Modal isOpen={false} onClose={mockOnClose}>
        <div data-testid="content">
          <h1>Modal content</h1>
        </div>
      </Modal>
    );
    const modalOverlay = queryByTestId("modal-overlay");
    const modalContent = queryByTestId("modal-content");
    const modalChildren = queryByTestId("content");
    await waitFor(() => {
      expect(mockOnClose).not.toBeCalled();
      expect(modalOverlay).toBeNull();
      expect(modalContent).toBeNull();
      expect(modalChildren).toBeNull();
    });
  });
});
