import {
  mockedDCAPolicy,
  mockPluginPolicy,
} from "@/__tests__/utils/global-mocks";
import PolicyTable from "@/modules/policy/components/policy-table/PolicyTable";
import { render } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import * as reactTable from "@tanstack/react-table";
import * as policyUtil from "@/modules/policy/utils/policy.util";

const mockPolicySchemaMap = new Map().set("dca", mockedDCAPolicy);
const mockPolicyMap = new Map().set("", mockPluginPolicy);

vi.mock("@/modules/policy/context/PolicyProvider", () => ({
  POLICY_ITEMS_PER_PAGE: 1,
  usePolicies: () => ({
    policyMap: mockPolicyMap,
    policySchemaMap: mockPolicySchemaMap,
    pluginType: "dca",
  }),
}));

// Note: This is needed in order to allow us to use spyOn...
vi.mock("@tanstack/react-table", async (importActual) => ({
  __esModule: true,
  ...(await importActual()),
}));

const useReactTableSpy = vi.spyOn(reactTable, "useReactTable");
const mapTableColumnDataSpy = vi.spyOn(policyUtil, "mapTableColumnData");

describe("PolicyTable", () => {
  it("should render all key components", async () => {
    const { findByTestId, findAllByTestId } = render(<PolicyTable />);
    await findByTestId("policy-table");
    const tableHeaders = await findAllByTestId("policy-table-headers");
    const tableCells = await findAllByTestId("policy-table-cells");
    expect(tableHeaders.length).toBe(mockedDCAPolicy.table.columns.length + 1);
    expect(tableCells.length).toBe(mockedDCAPolicy.table.columns.length + 1);
  });
  it("should call all needed functions", async () => {
    render(<PolicyTable />);
    expect(useReactTableSpy).toBeCalled();
    expect(mapTableColumnDataSpy).toBeCalled();
  });
});
