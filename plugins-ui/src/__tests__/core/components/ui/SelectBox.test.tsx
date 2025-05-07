import SelectBox from "@/modules/core/components/ui/select-box/SelectBox";
import userEvent from "@testing-library/user-event";
import { render } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";

describe("SelectBox", () => {
  it("should call onSelectChange", async () => {
    const mockOnSelectChange = vi.fn();
    const { findByTestId, findAllByTestId } = render(
      <SelectBox
        options={["Item #1", "Item #2", "Item #3"]}
        onSelectChange={mockOnSelectChange}
      />
    );

    const trigger = await findByTestId("select-box-trigger");
    await userEvent.click(trigger);
    const optionsList = await findAllByTestId("select-box-option");
    expect(optionsList.length).toEqual(3);
    await userEvent.click(optionsList[1]);
    expect(mockOnSelectChange).toBeCalledWith("Item #2");
    const selectedItem = await findByTestId("select-box-selected");
    expect(selectedItem.innerHTML).toContain("Item #2");
  });
  it("should show/hide options", async () => {
    const mockOnSelectChange = vi.fn();
    const { findByTestId, queryAllByTestId } = render(
      <SelectBox
        options={["Item #1", "Item #2", "Item #3"]}
        onSelectChange={mockOnSelectChange}
      />
    );

    const trigger = await findByTestId("select-box-trigger");
    await userEvent.click(trigger);
    const optionsList = queryAllByTestId("select-box-option");
    expect(optionsList.length).toEqual(3);
    await userEvent.click(trigger);
    const hiddenOptions = queryAllByTestId("select-box-option");
    expect(hiddenOptions.length).toBe(0);
  });
});
