import {
  ColumnDef,
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { useEffect, useState } from "react";
import { usePolicies } from "@/modules/policy/context/PolicyProvider";
import PolicyFilters from "../policy-filters/PolicyFilters";
import "./PolicyTable.css";
import tableJson from "../../schema/tableSchema.json";
import TokenPair from "@/modules/shared/token-pair/TokenPair";
import PolicyActions from "../policy-actions/PolicyActions";
import TokenName from "@/modules/shared/token-name/TokenName";
import TokenAmount from "@/modules/shared/token-amount/TokenAmount";
import { mapTableColumnData } from "../../utils/policy.util";
import ActiveStatus from "@/modules/shared/active-status/ActiveStatus";

const componentMap: Record<string, React.FC<any>> = {
  TokenPair,
  TokenName,
  TokenAmount,
  ActiveStatus,
};

const columns: ColumnDef<any>[] = tableJson.columns.map((col) => {
  const column: ColumnDef<any> = {
    accessorKey: col.accessorKey,
    header: col.header,
  };

  if (col.cellComponent) {
    [
      (column.cell = ({ getValue }) => {
        const Component = componentMap[col.cellComponent];
        return Component ? <Component data={getValue()} /> : getValue();
      }),
    ];
  }

  return column;
});

// all policies must have these actions Pause/Play, Edit, Tx history, Delete
columns.push({
  header: "Actions",
  cell: (info: any) => {
    const policyId = info.row.original.policyId;
    return <PolicyActions policyId={policyId} />;
  },
});

const PolicyTable = () => {
  const [data, setData] = useState<any>(() => []);
  const { policyMap } = usePolicies();

  useEffect(() => {
    const transformedData = [];

    for (const [_, value] of policyMap) {
      const obj: Record<string, any> = mapTableColumnData(
        value,
        tableJson.mapping
      );

      transformedData.push(obj);
    }

    setData(transformedData);
  }, [policyMap]);

  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]); // can set initial column filter state here

  const table = useReactTable({
    data,
    columns,
    state: {
      columnFilters,
    },
    onColumnFiltersChange: setColumnFilters,
    getFilteredRowModel: getFilteredRowModel(), // needed for client-side filtering
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <div>
      <PolicyFilters onFiltersChange={setColumnFilters} />
      <table className="policy-table">
        <thead>
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <th key={header.id}>
                  {header.isPlaceholder
                    ? null
                    : flexRender(
                        header.column.columnDef.header,
                        header.getContext()
                      )}
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody>
          {table.getRowModel().rows.map((row) => (
            <tr key={row.id}>
              {row.getVisibleCells().map((cell) => (
                <td key={cell.id}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </td>
              ))}
            </tr>
          ))}
          {table.getRowModel().rows.length === 0 && (
            <tr>
              <td colSpan={table.getAllColumns().length}>
                Nothing to see here yet.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
};

export default PolicyTable;
