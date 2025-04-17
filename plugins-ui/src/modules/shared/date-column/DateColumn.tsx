interface IDateColumnProps {
  data: string;
}

const DateColumn = ({ data }: IDateColumnProps) => {
  const parsedDate = new Date(data || "").toUTCString();
  if (parsedDate === "Invalid Date") {
    return <span>N/A</span>;
  }
  return <span>{parsedDate}</span>;
};

export default DateColumn;
