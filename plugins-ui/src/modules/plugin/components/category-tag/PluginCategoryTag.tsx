import "./PluginCategoryTag.css";

type PluginCategoryTagProps = {
  label: string;
};

const getCategoryStyles = (label: string): React.CSSProperties => {
  switch (label) {
    case "AI Agents":
      return {
        color: "#2955df",
        backgroundColor: "#ffffff"
      };
    case "Plugin":
      return {
        color: "#000000",
        backgroundColor: "#4cdcbf"
      };
    default:
      return {
        color: "#2955df",
        backgroundColor: "#ffffff"
      };
  }
};

const PluginCategoryTag = ({ label }: PluginCategoryTagProps) => {
  return (
    <span className="plugin-category-tag" style={getCategoryStyles(label)}>
      {label}
    </span>
  );
};

export default PluginCategoryTag;
