import Button from "@/modules/core/components/ui/button/Button";
import SelectBox from "@/modules/core/components/ui/select-box/SelectBox";
import Grid from "@/assets/Grid.svg?react";
import List from "@/assets/List.svg?react";
import Search from "@/assets/Search.svg?react";
import "./MarketplaceFilters.css";
import { useState } from "react";
import { ViewFilter } from "../../models/marketplace";

type MarketplaceFiltersProps = {
  viewFilter: ViewFilter;
  onChange: (view: ViewFilter) => void;
};

const sortOptions = ["Date (ASC)"]
const DEFAULT_SORTING = "Date (ASC)"

const MarketplaceFilters = ({
  viewFilter,
  onChange,
}: MarketplaceFiltersProps) => {
  const [view, setView] = useState<ViewFilter>(viewFilter);
  const [search, setSearch] = useState("");

  const changeView = (view: ViewFilter) => {
    setView(view);
    onChange(view);
  };

  const handleSortingChange = () => {
    console.log('sorting change');
  }

  return (
    <div className="filters">
      <div className="search">
        <div className="search-input">
          <input
            id="plugin-search"
            name="search"
            type="text"
            placeholder="Search by ..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <Search className="icon" width="20px" height="20px" />
        </div>
      </div>
      <div className="sort">
        <SelectBox
          options={sortOptions}
          value={DEFAULT_SORTING}
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
