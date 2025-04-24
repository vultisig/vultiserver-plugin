import PluginCard from "@/modules/plugin/components/plugin-card/PluginCard";
import "./Marketplace.css";
import MarketplaceFilters from "../marketplace-filters/MarketplaceFilters";
import { PluginFilters } from "../marketplace-filters/MarketplaceFilters";
import { useEffect, useState } from "react";
import { PluginMap, ViewFilter } from "../../models/marketplace";
import { Category } from "../../models/category";
import Toast from "@/modules/core/components/ui/toast/Toast";
import MarketplaceService from "../../services/marketplaceService";
import Pagination from "@/modules/core/components/ui/pagination/Pagination";

const getSavedView = (): string => {
  return localStorage.getItem("view") || "grid";
};

const getCategoryName = (categories: Category[], id: string) => {
  const category = categories.find(c => c.id === id);
  if (!category) return "";

  return category.name;
};

const ITEMS_PER_PAGE = 6;
const DEBOUNCE_DELAY = 500;

const Marketplace = () => {
  const [view, setView] = useState<string>(getSavedView());

  const [currentPage, setCurrentPage] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [filters, setFilters] = useState<PluginFilters>({
    term: "",
    categoryId: "",
    sortBy: "created_at",
    sortOrder: "DESC"
  });
  const [categories, setCategories] = useState<Category[]>([]);

  const changeView = (view: ViewFilter) => {
    localStorage.setItem("view", view);
    setView(view);
  };

  const [toast, setToast] = useState<{
    message: string;
    error?: string;
    type: "success" | "error";
  } | null>(null);

  const [pluginsMap, setPlugins] = useState<PluginMap | null>(null);

  useEffect(() => {
    const fetchCategories = async (): Promise<void> => {
      try {
        const fetchedCategories = await MarketplaceService.getCategories();
        setCategories(fetchedCategories);
      } catch (error: any) {
        console.error("Failed to get categories:", error.message);
        setToast({
          message: "Failed to get categories",
          error: error.error,
          type: "error",
        });
      }
    };

    fetchCategories();
  }, []);

  useEffect(() => {
    const fetchPlugins = async (): Promise<void> => {
      try {
        const fetchedPlugins = await MarketplaceService.getPlugins(
          filters.term,
          filters.categoryId,
          filters.sortBy,
          filters.sortOrder,
          currentPage > 1 ? (currentPage - 1) * ITEMS_PER_PAGE : 0,
          ITEMS_PER_PAGE
        );
        setPlugins(fetchedPlugins);
        setTotalPages(Math.ceil(fetchedPlugins.total_count / ITEMS_PER_PAGE));

        if (
          fetchedPlugins.total_count / ITEMS_PER_PAGE > 1 &&
          currentPage === 0
        ) {
          setCurrentPage(1);
        }
      } catch (error: any) {
        console.error("Failed to get plugins:", error.message);
        setToast({
          message: "Failed to get plugins",
          error: error.error,
          type: "error",
        });
      }
    };

    const timeout = setTimeout(() => {
      fetchPlugins();
    }, DEBOUNCE_DELAY);

    return () => clearTimeout(timeout);
  }, [filters, currentPage]);

  const onCurrentPageChange = (page: number): void => {
    setCurrentPage(page);
  };

  return (
    <>
      {categories.length && pluginsMap && (
        <div className="only-section">
          <h2>Plugins Marketplace</h2>
          <MarketplaceFilters
            categories={categories}
            viewFilter={view as ViewFilter}
            onViewChange={changeView}
            filters={filters}
            onFiltersChange={setFilters}
          />
          <section className="cards">
            {pluginsMap.plugins?.map((plugin) => (
              <div
                className={view === "list" ? "list-card" : ""}
                key={plugin.id}
              >
                <PluginCard
                  uiStyle={view as ViewFilter}
                  id={plugin.id}
                  title={plugin.title}
                  description={plugin.description}
                  categoryName={getCategoryName(categories, plugin.category_id)}
                />
              </div>
            ))}
          </section>

          {totalPages > 1 && (
            <Pagination
              currentPage={currentPage}
              totalPages={totalPages}
              onPageChange={onCurrentPageChange}
            />
          )}
        </div>
      )}

      {toast && (
        <Toast
          title={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}
    </>
  );
};

export default Marketplace;
