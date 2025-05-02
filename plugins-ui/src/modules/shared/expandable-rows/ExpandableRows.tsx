import { Row } from "@tanstack/react-table";
import style from "./ExpandableRows.module.css";

export type ExpandableRowsProps = {
  row: Row<unknown>;
};

const ExpandableRows = ({ row }: ExpandableRowsProps) => {
  return (
    <button
      onClick={() => {
        row.toggleExpanded();
      }}
      className={style.expandButton}
    >
      <div className={style.expandButtonContainer}>
        View {row.subRows?.length} row{row.subRows?.length > 1 ? "s" : ""}{" "}
        <span
          className={`${style.expandButtonChevron} ${row.getIsExpanded() ? style.expandButtonChevronRotated : ""}`}
        />
      </div>
    </button>
  );
};

export default ExpandableRows;
