import Button from "@/modules/core/components/ui/button/Button";
import SelectBox from "@/modules/core/components/ui/select-box/SelectBox";
import Grid from "@/assets/Grid.svg?react";
import List from "@/assets/List.svg?react";
import Search from "@/assets/Search.svg?react";
import "./MarketplaceFilters.css";
import { useState } from "react";
import { ViewFilter } from "../../models/marketplace";
import { Category } from "../../models/category";

export type PluginFilters = {
  term: string;
  categoryId: string;
  sortBy: string;
  sortOrder: string;
};

type MarketplaceFiltersProps = {
  categories: Category[];
  viewFilter: ViewFilter;
  onViewChange: (view: ViewFilter) => void;
  filters: PluginFilters;
  onFiltersChange: (filters: PluginFilters) => void;
};

const getCategoryIdByName = (categories: Category[], name: string) => {
  const category = categories.find((c) => c.name === name);
  if (!category) return "";

  return category.id;
};

const sortOptions: { [key: string]: { field: string; order: string } } = {
  "Date (DESC)": {
    field: "created_at",
    order: "DESC",
  },
  "Date (ASC)": {
    field: "created_at",
    order: "ASC",
  },
};
const sortLabels = Object.keys(sortOptions);

const MarketplaceFilters = ({
  categories,
  viewFilter,
  onViewChange,
  filters,
  onFiltersChange,
}: MarketplaceFiltersProps) => {
  const categoryNames = categories.map((c) => c.name);
  const defaultCategoryName =
    (categories.find((c) => c.id === filters.categoryId) || {}).name || "";
  const defaultSorting =
    Object.keys(sortOptions).find((key) => {
      return (
        sortOptions[key].field === filters.sortBy &&
        sortOptions[key].order === filters.sortOrder
      );
    }) || "";

  const [view, setView] = useState<ViewFilter>(viewFilter);
  const [term, setTerm] = useState("");
  const [categoryName, setCategoryName] = useState(defaultCategoryName);
  const [sortLabel, setSortLabel] = useState(defaultSorting);

  const changeView = (view: ViewFilter) => {
    setView(view);
    onViewChange(view);
  };

  const handleSearchChange = (term: string) => {
    setTerm(term);
    onFiltersChange({
      term,
      categoryId: getCategoryIdByName(categories, categoryName),
      sortBy: sortOptions[sortLabel].field,
      sortOrder: sortOptions[sortLabel].order,
    });
  };

  const handleCategoryChange = (name: string) => {
    setCategoryName(name);
    onFiltersChange({
      term,
      categoryId: getCategoryIdByName(categories, name),
      sortBy: sortOptions[sortLabel].field,
      sortOrder: sortOptions[sortLabel].order,
    });
  };

  const handleSortingChange = (sortOption: string) => {
    setSortLabel(sortOption);
    onFiltersChange({
      term,
      categoryId: getCategoryIdByName(categories, categoryName),
      sortBy: sortOptions[sortOption].field,
      sortOrder: sortOptions[sortOption].order,
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
      <div className="select">
        <SelectBox
          label="Show:"
          options={categoryNames}
          value={categoryName}
          onSelectChange={handleCategoryChange}
        />
      </div>
      <div className="select">
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
        data-testid="marketplace-filters-grid"
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
        data-testid="marketplace-filters-list"
      >
        <List width="20px" height="20px" color="#F0F4FC" />
      </Button>
    </div>
  );
};

export default MarketplaceFilters;
