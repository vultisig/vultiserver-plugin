import Accordion from "@/modules/core/components/ui/accordion/Accordion";
import userEvent from "@testing-library/user-event";
import { render, waitFor, within } from "@testing-library/react";
import { CSSProperties } from "react";
import { describe, expect, it } from "vitest";

const getChildren = () => {
  return (
    <div>
      <span>Child #1</span>
      <span>Child #2</span>
      <span>Child #3</span>
    </div>
  );
};

const getHeader = () => {
  return <h1 data-testid="accordion-header">Header</h1>;
};

const getExpandButton = () => {
  return {
    text: "ExpandButton",
    style: { backgroundColor: "red" } as CSSProperties,
  };
};

const getRenderResult = () => {
  return render(
    <Accordion
      children={getChildren()}
      header={getHeader()}
      expandButton={getExpandButton()}
    />
  );
};

describe("Accordion", () => {
  it("should visualize expand button", async () => {
    const { findByTestId } = getRenderResult();
    const triggerButton = await findByTestId("accordion-trigger");
    const triggerButtonCss = triggerButton.getAttribute("style");
    expect(triggerButton).toBeVisible();
    expect(triggerButtonCss).toContain("background-color: red;");
  });

  it("should visualize children", async () => {
    const { findByTestId } = getRenderResult();
    const accordionWrapper = await findByTestId("accordion-wrapper");
    const triggerButton = await within(accordionWrapper).findByRole("button");
    userEvent.click(triggerButton);
    const accordionChildren = await findByTestId("accordion-children");
    waitFor(() => {
      expect(accordionChildren).toBeVisible();
    });
  });

  it("should visualize header", async () => {
    const { findByTestId } = getRenderResult();
    const accordionHeader = await findByTestId("accordion-header");
    expect(accordionHeader).toBeVisible();
  });
});
