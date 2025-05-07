import Button from "@/modules/core/components/ui/button/Button";
import { useNavigate } from "react-router-dom";
import logo from "../../../../assets/DCA-image.png"; // Adjust path based on file location
import "./PluginCard.css";
import { ViewFilter } from "@/modules/marketplace/models/marketplace";
import PluginCategoryTag from "@/modules/plugin/components/category-tag/PluginCategoryTag";

const truncateText = (text: string, maxLength: number = 500): string => {
  return text.length > maxLength ? text.slice(0, maxLength) + "..." : text;
};

type PluginCardProps = {
  id: string;
  title: string;
  description: string;
  uiStyle: ViewFilter;
  categoryName: string;
};

const PluginCard = ({
  id,
  uiStyle,
  title,
  description,
  categoryName,
}: PluginCardProps) => {
  const navigate = useNavigate();

  return (
    <div className={`plugin ${uiStyle}`} data-testid="plugin-card-wrapper">
      <div className={uiStyle === "grid" ? "" : "info-group"}>
        <img data-testid="plugin-card-logo" src={logo} alt={title} />

        <div className="plugin-info">
          <PluginCategoryTag label={categoryName} />
          <h3 data-testid="plugin-card-title">{title}</h3>
          <p data-testid="plugin-card-description">
            {truncateText(description)}
          </p>
        </div>
      </div>

      <Button
        style={uiStyle === "grid" ? { width: "100%" } : { minWidth: "95px" }}
        size={uiStyle === "grid" ? "small" : "mini"}
        type="button"
        styleType="primary"
        onClick={() => navigate(`/plugins/${id}`)}
        data-testid="plugin-card-details-btn"
      >
        See details
      </Button>
    </div>
  );
};

export default PluginCard;
