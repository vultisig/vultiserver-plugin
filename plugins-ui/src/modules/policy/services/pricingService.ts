import { get, post } from "@/modules/core/services/httpService";
import { PluginPricing } from "../models/pluginPricing";

const getPublicKey = () => localStorage.getItem("publicKey");

const PricingService = {
  getPluginPricing: async (pluginUrl: string, pluginType: string) => {
    try {
      const endpoint = `${pluginUrl}/plugin/${encodeURIComponent(pluginType)}/pricing-policy`;
      const pluginPricing = await get(endpoint, {
        headers: {
          plugin_type: pluginType,
          public_key: getPublicKey(),
          Authorization: `Bearer ${localStorage.getItem("authToken")}`,
        },
      });
      return pluginPricing;
    } catch (error) {
      console.error("Error getting pricing:", error);
      throw error;
    }
  },

  createPricing: async (pluginUrl: string, pricing: Omit<PluginPricing, "id">) => {
    try {
      const endpoint = `${pluginUrl}/plugin/${encodeURIComponent(pricing.plugin_type)}/pricing-policy`;
      const newPricing = await post(endpoint, pricing);
      return newPricing;
    } catch (error) {
      console.error("Error creating pricing:", error);
      throw error;
    }
  }
}

export default PricingService;
