import Summary, { type Row } from "@/modules/shared/summary/Summary";
import { render } from "@testing-library/react";
import { describe, it } from "vitest";

const mockData: Row[] = [];

describe("Summary", () => {
  it("should render", async () => {
    const { findByTestId } = render(<Summary title="Test" data={mockData} />);
    await findByTestId("accordion-wrapper");
  });
});
