import TransactionHistory from "@/modules/policy/components/transaction-history/TransactionHistory";
import { render, waitFor, within } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { mockEventBus } from "@/__tests__/utils/global-mocks";

const mockDate = new Date().toString();

const hoisted = vi.hoisted(() => ({
  mockGetPolicyHistory: vi.fn(() => ({
    history: [{ id: "1", updated_at: mockDate, status: "active" }],
    total_count: 1,
  })),
}));

vi.mock("@/modules/policy/context/PolicyProvider", () => ({
  POLICY_ITEMS_PER_PAGE: 1,
  usePolicies: () => ({
    getPolicyHistory: hoisted.mockGetPolicyHistory,
  }),
}));

describe("TransactionHistory", () => {
  it("should render all components and call all functions", async () => {
    const { findByTestId, findAllByTestId } = render(
      <TransactionHistory policyId="1" />
    );
    await findByTestId("transaction-history-wrapper");
    const transactionHistoryItems = await findAllByTestId(
      "transaction-history-entries"
    );
    const status = await within(transactionHistoryItems[0]).findByTestId(
      "status"
    );
    const date = await within(transactionHistoryItems[0]).findByTestId("date");
    const time = await within(transactionHistoryItems[0]).findByTestId("time");
    expect(transactionHistoryItems.length).toBe(1);
    expect(status).toHaveTextContent("active");
    expect(date).toHaveTextContent(
      new Date(mockDate).toLocaleDateString("en-GB", {
        day: "2-digit",
        month: "short",
        year: "numeric",
      })
    );
    expect(time).toHaveTextContent(
      new Date(mockDate).toLocaleTimeString("en-GB", {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
      })
    );
    expect(hoisted.mockGetPolicyHistory).toBeCalledWith("1", 0, 25);
  });

  it("should show toast on error", async () => {
    hoisted.mockGetPolicyHistory.mockRejectedValueOnce(new Error("From tests"));
    const { findByTestId } = render(<TransactionHistory policyId="1" />);
    await findByTestId("transaction-history-no-results");
    await waitFor(() => {
      expect(mockEventBus.publish).toBeCalledWith("onToast", {
        message: "From tests",
        type: "error",
      });
    });
  });
});
