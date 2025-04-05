import Button from "@/modules/core/components/ui/button/Button";
import SelectBox from "@/modules/core/components/ui/select-box/SelectBox";
import Grid from "@/assets/Grid.svg?react";
import List from "@/assets/List.svg?react";
import Search from "@/assets/Search.svg?react";
import "./MarketplaceFilters.css";
import { useState } from "react";
import { ViewFilter } from "../../models/marketplace";

export type PluginFilters = {
  term: string;
  sortBy: string;
  sortOrder: string;
};

type MarketplaceFiltersProps = {
  viewFilter: ViewFilter;
  onViewChange: (view: ViewFilter) => void;
  onFilterChange: (filters: PluginFilters) => void;
};

const sortLabels = ["Date (DESC)", "Date (ASC)"];
const sortOptions: { [key: string]: { field: string, order: string } } = {
  "Date (ASC)": {
    field: "created_at",
    order: "ASC"
  },
  "Date (DESC)": {
    field: "created_at",
    order: "DESC"
  }
};
const DEFAULT_SORTING = sortLabels[0];

const MarketplaceFilters = ({
  viewFilter,
  onViewChange,
  onFilterChange,
}: MarketplaceFiltersProps) => {
  const [view, setView] = useState<ViewFilter>(viewFilter);
  const [term, setTerm] = useState("");
  const [sortLabel, setSortLabel] = useState(DEFAULT_SORTING);

  const changeView = (view: ViewFilter) => {
    setView(view);
    onViewChange(view);
  };

  const handleSearchChange = (term: string) => {
    setTerm(term);
    onFilterChange({
      term,
      sortBy: sortOptions[sortLabel].field,
      sortOrder: sortOptions[sortLabel].order
    });
  };

  const handleSortingChange = (sortOption: string) => {
    setSortLabel(sortOption);
    onFilterChange({
      term,
      sortBy: sortOptions[sortOption].field,
      sortOrder: sortOptions[sortOption].order
    });
  };

  return (
    <div className="filters">
      <div className="search">
        <div className="search-input">
          <input
            id="plugin-search"
            name="search"
            type="text"
            placeholder="Search by ..."
            value={term}
            onChange={(e) => handleSearchChange(e.target.value)}
          />
          <Search className="icon" width="20px" height="20px" />
        </div>
      </div>
      <div className="sort">
        <SelectBox
          options={sortLabels}
          value={sortLabel}
          onSelectChange={handleSortingChange}
        />
      </div>
      <Button
        ariaLabel="Grid view"
        type="button"
        styleType="tertiary"
        size="medium"
        className={`view-filter ${view === "grid" ? "active" : ""}`}
        onClick={() => changeView("grid")}
      >
        <Grid width="20px" height="20px" color="#F0F4FC" />
      </Button>
      <Button
        ariaLabel="List view"
        type="button"
        styleType="tertiary"
        size="medium"
        className={`view-filter ${view === "list" ? "active" : ""}`}
        onClick={() => changeView("list")}
      >
        <List width="20px" height="20px" color="#F0F4FC" />
      </Button>
    </div>
  );
};

export default MarketplaceFilters;
