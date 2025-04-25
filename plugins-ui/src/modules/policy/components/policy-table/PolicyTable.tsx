import {
  ColumnDef,
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { useEffect, useState } from "react";
import {
  POLICY_ITEMS_PER_PAGE,
  usePolicies,
} from "@/modules/policy/context/PolicyProvider";
import PolicyFilters from "../policy-filters/PolicyFilters";
import "./PolicyTable.css";
import TokenPair from "@/modules/shared/token-pair/TokenPair";
import PolicyActions from "../policy-actions/PolicyActions";
import TokenName from "@/modules/shared/token-name/TokenName";
import TokenAmount from "@/modules/shared/token-amount/TokenAmount";
import { mapTableColumnData } from "../../utils/policy.util";
import ActiveStatus from "@/modules/shared/active-status/ActiveStatus";
import { PolicySchema } from "../../models/policy";
import Pagination from "@/modules/core/components/ui/pagination/Pagination";

const componentMap: Record<string, React.FC<any>> = {
  TokenPair,
  TokenName,
  TokenAmount,
  ActiveStatus,
};

const getTableColumns = (schema: PolicySchema) => {
  const columns: ColumnDef<any>[] = schema.table.columns.map((col: any) => {
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

  return columns;
};

const PolicyTable = () => {
  const [data, setData] = useState<any>(() => []);
  const {
    policyMap,
    policySchemaMap,
    pluginType,
    policiesTotalCount,
    currentPage,
    setCurrentPage,
  } = usePolicies();
  const [columns, setColumns] = useState<ColumnDef<any>[]>([]);
  const [totalPages, setTotalPages] = useState(0);

  useEffect(() => {
    const savedSchema = policySchemaMap.get(pluginType);

    if (
      savedSchema &&
      savedSchema.table &&
      savedSchema.table.columns &&
      savedSchema.table.mapping
    ) {
      const mappedColumns: ColumnDef<any>[] = getTableColumns(savedSchema);

      setColumns(mappedColumns);

      const transformedData = [];
      for (const [_, value] of policyMap) {
        const obj: Record<string, any> = mapTableColumnData(
          value,
          savedSchema.table.mapping
        );
        transformedData.push(obj);
      }
      setData(transformedData);
    }
  }, [policySchemaMap, policyMap]);

  useEffect(() => {
    setTotalPages(Math.ceil(policiesTotalCount / POLICY_ITEMS_PER_PAGE));
    if (policiesTotalCount / POLICY_ITEMS_PER_PAGE > 1 && currentPage === 0) {
      setCurrentPage(1);
    }
  }, [currentPage, policiesTotalCount]);

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

  const onCurrentPageChange = (page: number): void => {
    setCurrentPage(page);
  };

  if (columns.length === 0) return;

  return (
    <div>
      <PolicyFilters onFiltersChange={setColumnFilters} />

      {policySchemaMap.has(pluginType) && (
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
                <td
                  colSpan={table.getAllColumns().length}
                  className="empty-message-row"
                >
                  Nothing to see here yet.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      )}

      {totalPages > 1 && (
        <Pagination
          currentPage={currentPage}
          totalPages={totalPages}
          onPageChange={onCurrentPageChange}
        />
      )}
    </div>
  );
};

export default PolicyTable;
