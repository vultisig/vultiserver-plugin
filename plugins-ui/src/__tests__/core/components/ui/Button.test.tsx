import Button from "@/modules/core/components/ui/button/Button";
import userEvent from "@testing-library/user-event";

import { render, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

describe("Button", () => {
  it("should call provided function on click", async () => {
    const mockOnClick = vi.fn();
    const { findByRole } = render(
      <Button
        size="medium"
        styleType="primary"
        type="button"
        className="test"
        onClick={mockOnClick}
      />
    );
    const button = await findByRole("button");
    userEvent.click(button);
    const buttonType = button.getAttribute("type");
    const buttonClasslist = button.classList;

    await waitFor(() => {
      expect(mockOnClick).toBeCalled();
      expect(buttonClasslist.contains("test")).toEqual(true);
      expect(buttonClasslist.contains("primary")).toEqual(true);
      expect(buttonClasslist.contains("medium")).toEqual(true);
      expect(buttonType).toEqual("button");
    });
  });
});
