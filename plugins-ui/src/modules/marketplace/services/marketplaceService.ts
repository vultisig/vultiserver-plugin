import { get, post } from "@/modules/core/services/httpService";
import { PluginMap } from "../models/marketplace";
import { Category } from "../models/category";
import { Plugin } from "@/modules/plugin/models/plugin";
import {
  PluginPolicy,
  PolicyTransactionHistory,
} from "@/modules/policy/models/policy";

const getPublicKey = () => localStorage.getItem("publicKey");
const getMarketplaceUrl = () => import.meta.env.VITE_MARKETPLACE_URL;

const MarketplaceService = {
  /**
   * Get plugins from the API.
   * @returns {Promise<Object>} A promise that resolves to the fetched plugins.
   */
  getPlugins: async (
    term: string,
    categoryId: string,
    sortBy: string,
    sortOrder: string,
    skip: number,
    take: number
  ): Promise<PluginMap> => {
    try {
      const sort = sortOrder === "DESC" ? `-${sortBy}` : sortBy
      const endpoint = `${getMarketplaceUrl()}/plugins?term=${
        encodeURIComponent(term)
      }&category_id=${
        encodeURIComponent(categoryId)
      }&sort=${
        encodeURIComponent(sort)
      }&skip=${skip}&take=${take}`;
      const plugins = await get(endpoint);
      return plugins;
    } catch (error) {
      console.error("Error getting plugins:", error);
      throw error;
    }
  },

  /**
   * Get all plugin categories
   * @returns {Promise<Object>} A promise that resolves to the fetched categories.
   */
  getCategories: async (): Promise<Category[]> => {
    try {
      const endpoint = `${getMarketplaceUrl()}/categories`;
      const categories = await get(endpoint);
      return categories;
    } catch (error) {
      console.error("Error getting categories:", error);
      throw error;
    }
  },

  /**
   * Get plugin by id from the API.
   * @returns {Promise<Object>} A promise that resolves to the fetched plugin.
   */
  getPlugin: async (id: string): Promise<Plugin> => {
    try {
      const endpoint = `${getMarketplaceUrl()}/plugins/${id}`;
      const plugin = await get(endpoint);
      return plugin;
    } catch (error) {
      console.error("Error getting plugin:", error);
      throw error;
    }
  },

  /**
   * Post signature, publicKey, chainCodeHex, derivePath to the APi
   * @returns {Promise<Object>} A promise that resolves with auth token.
   */
  getAuthToken: async (
    message: string,
    signature: string,
    publicKey: string,
    chainCodeHex: string,
    derivePath: string
  ): Promise<string> => {
    try {
      const endpoint = `${getMarketplaceUrl()}/auth`;
      const response = await post(endpoint, {
        message: message,
        signature: signature,
        public_key: publicKey,
        chain_code_hex: chainCodeHex,
        derive_path: derivePath,
      });
      return response.token;
    } catch (error) {
      console.error("Failed to get auth token", error);
      throw error;
    }
  },

  /**
   * Get policies from the API.
   * @returns {Promise<Object>} A promise that resolves to the fetched policies.
   */
  getPolicies: async (pluginType: string): Promise<PluginPolicy[]> => {
    try {
      const endpoint = `${getMarketplaceUrl()}/plugins/policies`;
      const newPolicy = await get(endpoint, {
        headers: {
          plugin_type: pluginType,
          public_key: getPublicKey(),
          Authorization: `Bearer ${localStorage.getItem("authToken")}`,
        },
      });
      return newPolicy;
    } catch (error: any) {
      if (error.message === "Unauthorized") {
        localStorage.removeItem("authToken");
        // Dispatch custom event to notify other components
        window.dispatchEvent(new Event("storage"));
      }
      console.error("Error getting policies:", error);
      throw error;
    }
  },

  /**
   * Get policy transaction history from the API.
   * @returns {Promise<Object>} A promise that resolves to the fetched policies.
   */
  getPolicyTransactionHistory: async (
    policyId: string
  ): Promise<PolicyTransactionHistory[]> => {
    try {
      const endpoint = `${getMarketplaceUrl()}/plugins/policies/${policyId}/history`;
      const history = await get(endpoint, {
        headers: {
          public_key: getPublicKey(),
          Authorization: `Bearer ${localStorage.getItem("authToken")}`,
        },
      });
      return history;
    } catch (error) {
      console.error("Error getting policy history:", error);

      throw error;
    }
  },
};

export default MarketplaceService;
