import {
  CellContext,
  ColumnFiltersState,
  ExpandedState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  Row,
  RowData,
  useReactTable,
} from "@tanstack/react-table";
import { useEffect, useState, Fragment } from "react";
import {
  POLICY_ITEMS_PER_PAGE,
  usePolicies,
} from "@/modules/policy/context/PolicyProvider";
import PolicyFilters from "@/modules/policy/components/policy-filters/PolicyFilters";
import "@/modules/policy/components/policy-table/PolicyTable.css";
import TokenPair, {
  type TokenPairProps,
} from "@/modules/shared/token-pair/TokenPair";
import PolicyActions from "../policy-actions/PolicyActions";
import TokenName, {
  type TokenNameProps,
} from "@/modules/shared/token-name/TokenName";
import TokenAmount, {
  type TokenAmountProps,
} from "@/modules/shared/token-amount/TokenAmount";
import { mapTableColumnData } from "../../utils/policy.util";
import ActiveStatus, {
  type ActiveStatusProps,
} from "@/modules/shared/active-status/ActiveStatus";
import DateColumn, {
  type DateColumnProps,
} from "@/modules/shared/date-column/DateColumn";
import TokenImage, {
  type TokenImageProps,
} from "@/modules/shared/token-image/TokenImage";
import { PolicySchema, PolicyTableColumn } from "../../models/policy";
import ExpandableRows, {
  type ExpandableRowsProps,
} from "@/modules/shared/expandable-rows/ExpandableRows";
import Pagination from "@/modules/core/components/ui/pagination/Pagination";

const componentMap: Record<
  string,
  ({ data, row }: { data: unknown; row: Row<unknown> }) => JSX.Element
> = {
  TokenPair: (props) => <TokenPair {...(props as TokenPairProps)} />,
  TokenName: (props) => <TokenName {...(props as TokenNameProps)} />,
  TokenAmount: (props) => <TokenAmount {...(props as TokenAmountProps)} />,
  ActiveStatus: (props) => <ActiveStatus {...(props as ActiveStatusProps)} />,
  DateColumn: (props) => <DateColumn {...(props as DateColumnProps)} />,
  TokenImage: (props) => <TokenImage {...(props as TokenImageProps)} />,
  ExpandableRows: (props) => (
    <ExpandableRows {...(props as ExpandableRowsProps)} />
  ),
};

const getTableColumns = (schema: PolicySchema) => {
  const columns: PolicyTableColumn[] = schema.table.columns.map(
    (col: PolicyTableColumn) => {
      const column: PolicyTableColumn = {
        accessorKey: col.accessorKey,
        header: col.header,
        expandable: col.expandable ?? false,
      };
      if (col.cellComponent) {
        column.cell = ({ getValue, row }) => {
          const func = componentMap[`${col.cellComponent}`];
          return func ? func({ data: getValue(), row }) : getValue();
        };
      }
      return column;
    }
  );

  // all policies must have these actions Pause/Play, Edit, Tx history, Delete
  columns.push({
    header: "Actions",
    cell: (info: CellContext<RowData, unknown>) => {
      const policyId = (info.row.original as Record<string, unknown>).policyId;
      return <PolicyActions policyId={`${policyId}`} />;
    },
  } as PolicyTableColumn);

  return columns;
};

const PolicyTable = () => {
  const [data, setData] = useState<unknown[]>([]);
  const {
    policyMap,
    policySchemaMap,
    pluginType,
    policiesTotalCount,
    currentPage,
    setCurrentPage,
  } = usePolicies();
  const [columns, setColumns] = useState<PolicyTableColumn[]>([]);
  const [totalPages, setTotalPages] = useState(0);

  useEffect(() => {
    const savedSchema = policySchemaMap.get(pluginType);

    if (
      savedSchema &&
      savedSchema.table &&
      savedSchema.table.columns &&
      savedSchema.table.mapping
    ) {
      const mappedColumns: PolicyTableColumn[] = getTableColumns(savedSchema);

      setColumns(mappedColumns);

      const transformedData = [];
      for (const [, value] of policyMap) {
        const obj: Record<string, unknown> = mapTableColumnData(
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

  const [expanded, setExpanded] = useState<ExpandedState>({});

  const table = useReactTable({
    data,
    columns,
    state: {
      columnFilters,
      expanded,
    },
    getSubRows: (row) => (row as Row<unknown>).subRows,
    onExpandedChange: setExpanded,
    onColumnFiltersChange: setColumnFilters,
    getFilteredRowModel: getFilteredRowModel(), // needed for client-side filtering
    getCoreRowModel: getCoreRowModel(),
  });

  const onCurrentPageChange = (page: number): void => {
    setCurrentPage(page);
  };

  if (columns.length === 0) return;

  const generateExpandableRows = (tableRow: Row<unknown>) => {
    const expandableRowData = tableRow.getValue("subRows") as Record<
      string,
      unknown
    >[];

    return expandableRowData.map((row, i) => {
      return (
        <p key={i}>
          {Object.keys(row).map((key, j) => {
            return (
              <Fragment key={j}>
                <span className="expandable-row-key">{key}:</span>
                <span>{`${row[key]}`} </span>
              </Fragment>
            );
          })}
        </p>
      );
    });
  };

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
            {table.getRowModel().rows.map((row: Row<unknown>, i) => (
              <Fragment key={`${i}-container`}>
                <tr key={i}>
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </td>
                  ))}
                </tr>
                {row.getIsExpanded() && (
                  <tr key={`${i}-expanded`}>
                    <td
                      key={`${i}-expanded-col`}
                      colSpan={columns.length}
                      style={{ paddingLeft: "30px" }}
                    >
                      {generateExpandableRows(row)}
                    </td>
                  </tr>
                )}
              </Fragment>
            ))}
            {table.getRowModel().rows.length === 0 && (
              <tr className="expandable-row">
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
