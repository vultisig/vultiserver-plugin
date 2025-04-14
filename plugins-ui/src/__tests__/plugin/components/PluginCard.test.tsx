import { ViewFilter } from "@/modules/marketplace/models/marketplace";
import PluginCard from "@/modules/plugin/components/plugin-card/PluginCard";
import { render } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";

const hoisted = vi.hoisted(() => ({
  mockUseNavigate: vi.fn(),
}));

vi.mock("react-router-dom", async (importActual) => ({
  ...(await importActual()),
  useNavigate: vi.fn(() => hoisted.mockUseNavigate),
}));

const mockPluginCardDetails = {
  id: "Id",
  title: "Title",
  description:
    "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu",
  uiStyle: "list" as ViewFilter,
};

describe("PluginCard", () => {
  it("should visualize all important details and call navigate", async () => {
    const { findByTestId } = render(<PluginCard {...mockPluginCardDetails} />);
    const detailsBtn = await findByTestId("plugin-card-details-btn");
    await userEvent.click(detailsBtn);
    expect((await findByTestId("plugin-card-title")).innerHTML).toEqual(
      mockPluginCardDetails.title
    );
    expect((await findByTestId("plugin-card-description")).innerHTML).toEqual(
      `${mockPluginCardDetails.description.slice(0, 500)}...`
    );
    expect(hoisted.mockUseNavigate).toBeCalledWith(
      `/plugins/${mockPluginCardDetails.id}`
    );
  });
  it("should add correct class based on uiStyle", async () => {
    const { findByTestId } = render(
      <PluginCard {...mockPluginCardDetails} uiStyle="grid" />
    );
    const pluginCardWrapper = await findByTestId("plugin-card-wrapper");
    expect(pluginCardWrapper.classList.contains("grid")).toBeTruthy();
  });
});
