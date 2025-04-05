import { get } from "@/modules/core/services/httpService";
import { PluginMap } from "../models/marketplace";

const getMarketplaceUrl = () => import.meta.env.VITE_MARKETPLACE_URL;

const MarketplaceService = {
  /**
   * Get plugins from the API.
   * @returns {Promise<Object>} A promise that resolves to the fetched plugins.
   */
  getPlugins: async (
    term: string,
    sortBy: string,
    sortOrder: string,
    skip: number,
    take: number
  ): Promise<PluginMap> => {
    try {
      const sort = sortOrder === "DESC" ? `-${sortBy}` : sortBy
      const endpoint = `${getMarketplaceUrl()}/plugins?term=${encodeURIComponent(term)}&sort=${encodeURIComponent(sort)}&skip=${skip}&take=${take}`;
      const plugins = await get(endpoint);
      return plugins;
    } catch (error) {
      console.error("Error getting plugins:", error);
      throw error;
    }
  },
};

export default MarketplaceService;
