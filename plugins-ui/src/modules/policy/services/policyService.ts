import { post, get, put, remove } from "@/modules/core/services/httpService";
import { PluginPolicy } from "../models/policy";

const PolicyService = {
  /**
   * Posts a new policy to the API.
   * @param {PluginPolicy} pluginPolicy - The policy to be created.
   * @returns {Promise<Object>} A promise that resolves to the created policy.
   */
  createPolicy: async (
    serverEndpoint: string,
    pluginPolicy: PluginPolicy
  ): Promise<PluginPolicy> => {
    try {
      const endpoint = `${serverEndpoint}/plugin/policy`;
      const newPolicy = await post(endpoint, pluginPolicy);
      return newPolicy;
    } catch (error) {
      console.error("Error creating policy:", error);
      throw error;
    }
  },

  /**
   * Updates policy to the API.
   * @param {PluginPolicy} pluginPolicy - The policy to be created.
   * @returns {Promise<Object>} A promise that resolves to the created policy.
   */
  updatePolicy: async (
    serverEndpoint: string,
    pluginPolicy: PluginPolicy
  ): Promise<PluginPolicy> => {
    try {
      const endpoint = `${serverEndpoint}/plugin/policy`;
      const newPolicy = await put(endpoint, pluginPolicy);
      return newPolicy;
    } catch (error) {
      console.error("Error updating policy:", error);
      throw error;
    }
  },

  /**
   * Delete policy from the API.
   * @param {id} string - The policy to be deleted.
   */
  deletePolicy: async (
    serverEndpoint: string,
    id: string,
    signature: string
  ) => {
    try {
      const endpoint = `${serverEndpoint}/plugin/policy/${id}`;
      return await remove(endpoint, { signature: signature });
    } catch (error) {
      console.error("Error deleting policy:", error);
      throw error;
    }
  },

  /**
   * Get PolicySchema
   * @returns {Promise<Object>} A promise that resolves to the fetched schema.
   */
  getPolicySchema: async (
    serverEndpoint: string,
    pluginType: string
  ): Promise<any> => {
    try {
      const endpoint = `${serverEndpoint}/plugin/policy/schema`;
      const newPolicy = await get(endpoint, {
        headers: {
          plugin_type: pluginType,
        },
      });
      return newPolicy;
    } catch (error) {
      console.error("Error getting policy schema:", error);
      throw error;
    }
  },
};

export default PolicyService;
