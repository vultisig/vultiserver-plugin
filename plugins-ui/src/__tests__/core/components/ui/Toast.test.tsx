import Toast from "@/modules/core/components/ui/toast/Toast";
import userEvent from "@testing-library/user-event";
import { render, waitFor, within } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { publish } from "@/utils/eventBus";
import { act } from "react";

vi.unmock("@/utils/eventBus");
describe("Toast", () => {
  it("should show toasts", async () => {
    const { findAllByTestId } = render(<Toast />);
    act(() => {
      publish("onToast", {
        type: "success",
        message: "Great success",
      });
      publish("onToast", {
        type: "success",
        message: "Great success",
      });
    });
    const toasts = await findAllByTestId("toast-item");
    expect(toasts.length).toEqual(2);
  });
  it("should close toast", async () => {
    const { findAllByTestId } = render(<Toast />);
    act(() => {
      publish("onToast", {
        type: "success",
        message: "Great success",
        duration: 1000,
      });
      publish("onToast", {
        type: "success",
        message: "Great success",
        duration: 1000,
      });
    });
    const toasts = await findAllByTestId("toast-item");
    await waitFor(async () => {
      expect(toasts.length).toEqual(2);
      const closeTrigger = await within(toasts[0]).findByRole("button");
      await userEvent.click(closeTrigger);
      const updatedToasts = await findAllByTestId("toast-item");
      expect(updatedToasts.length).toEqual(1);
    });
  });
});
