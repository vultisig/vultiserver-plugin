export type DateColumnProps = {
  data: string;
};

const DateColumn = ({ data }: DateColumnProps) => {
  const parsedDate = new Date(data || "").toUTCString();
  if (parsedDate === "Invalid Date") {
    return <span>N/A</span>;
  }
  return <span>{parsedDate}</span>;
};

export default DateColumn;
