import { mockEventBus } from "@/__tests__/utils/global-mocks";
import PluginDetail from "@/modules/plugin/components/plugin-detail/PluginDetail";
import { Plugin } from "@/modules/plugin/models/plugin";
import { render, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";

const hoisted = vi.hoisted(() => ({
  mockUseNavigate: vi.fn(),
  mockGetPlugin: vi.fn(),
}));
const pluginId = "Plugin ID";
const mockPlugin = {
  id: pluginId,
  type: "Type",
  title: "Title",
  description: "Description",
  metadata: {},
  server_endpoint: "Server endpoint",
  pricing_id: "Pricing ID",
  ratings: [
    {
      rating: 1,
      count: 1,
    },
  ],
} as Plugin;

vi.mock("react-router-dom", async (importActual) => ({
  ...(await importActual()),
  useNavigate: vi.fn(() => hoisted.mockUseNavigate),
  useParams: vi.fn(() => ({ pluginId })),
}));

vi.mock("@/modules/marketplace/services/marketplaceService", () => ({
  default: {
    getPlugin: hoisted.mockGetPlugin,
    getReviews: vi.fn(() => ({
      reviews: [],
      total_count: 0,
    })),
  },
}));

describe("PluginDetail", () => {
  it("should render all important components call navigate to install plugin", async () => {
    hoisted.mockGetPlugin.mockReturnValueOnce(mockPlugin);
    const { findByTestId } = render(<PluginDetail />);
    const installBtn = await findByTestId("plugin-detail-install-btn");
    await userEvent.click(installBtn);

    await findByTestId("plugin-detail-back-btn");
    await findByTestId("leave-review-wrapper");
    await findByTestId("rating-wrapper");
    await findByTestId("review-history-wrapper");

    expect(hoisted.mockUseNavigate).toBeCalledWith(
      `/plugins/${pluginId}/policies`
    );
  });
  it("should handle error from getPlugin", async () => {
    hoisted.mockGetPlugin.mockRejectedValueOnce(new Error("From test"));
    const { queryByTestId } = render(<PluginDetail />);
    const installBtn = queryByTestId("plugin-detail-install-btn");
    await waitFor(() => {
      expect(installBtn).toBeNull();
      expect(mockEventBus.publish).toBeCalledWith("onToast", {
        message: "Failed to get plugin",
        type: "error",
      });
    });
  });
  it("should call navigate when clicking 'Back to All Plugins'", async () => {
    const { findByTestId } = render(<PluginDetail />);
    const backBtn = await findByTestId("plugin-detail-back-btn");
    await userEvent.click(backBtn);
    expect(hoisted.mockUseNavigate).toBeCalledWith("/plugins");
  });
});
