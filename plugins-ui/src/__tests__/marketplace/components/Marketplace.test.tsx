import { mockEventBus } from "@/__tests__/utils/global-mocks";
import Marketplace from "@/modules/marketplace/components/marketplace-main/Marketplace";
import { render, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi, afterEach } from "vitest";

const hoisted = vi.hoisted(() => ({
  mockGetPlugins: vi.fn(() => {
    return {
      plugins: [
        { id: "Test #1", title: "Test #1", description: "Test plugin" },
        { id: "Test #2", title: "Test #2", description: "Test plugin" },
      ],
      total_count: 20,
    };
  }),
  mockGetCategories: vi.fn(() => [{ id: "1", name: "Test Category" }]),
}));

vi.mock("@/modules/marketplace/services/marketplaceService", async () => ({
  default: {
    getPlugins: hoisted.mockGetPlugins,
    getCategories: hoisted.mockGetCategories,
  },
}));

vi.mock("react-router-dom", async (importActual) => ({
  ...(await importActual()),
  useNavigate: vi.fn(),
}));

describe("Marketplace", () => {
  afterEach(() => {
    vi.resetAllMocks();
  });
  it("should fetch plugins", async () => {
    const { findByTestId, findAllByTestId } = render(<Marketplace />);

    await findByTestId("marketplace-filters-grid");
    await findByTestId("pagination-wrapper");
    const pluginCards = await findAllByTestId("plugin-card-wrapper");
    expect(hoisted.mockGetPlugins).toBeCalled();
    expect(pluginCards.length).toEqual(2);
  });
  it("should show toast when getPlugins fails", async () => {
    hoisted.mockGetPlugins.mockRejectedValueOnce(new Error("should throw"));
    render(<Marketplace />);

    await waitFor(() => {
      expect(hoisted.mockGetPlugins).toBeCalled();
      expect(mockEventBus.publish).toBeCalledWith("onToast", {
        type: "error",
        message: "Failed to get plugins",
      });
    });
  });
  it("should update layout", async () => {
    const { findByTestId, findAllByTestId } = render(<Marketplace />);
    const filtersListButton = await findByTestId("marketplace-filters-list");
    await userEvent.click(filtersListButton);
    await waitFor(async () => {
      const pluginCards = await findAllByTestId("marketplace-plugin-card");
      const results = pluginCards.map((card) =>
        card.classList.contains("list-card")
      );
      expect(results).toStrictEqual([true, true]);
    });
  });
  it("should change page", async () => {
    const { findAllByTestId } = render(<Marketplace />);
    const pages = await findAllByTestId("pagination-page");
    await userEvent.click(pages[2]);
    await waitFor(async () => {
      expect(hoisted.mockGetPlugins).toBeCalledWith(
        "",
        "",
        "created_at",
        "DESC",
        12,
        6
      );
    });
  });
});
