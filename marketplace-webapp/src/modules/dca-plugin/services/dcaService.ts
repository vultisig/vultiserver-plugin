import { post } from "@/modules/core/services/httpService";
import { Policy } from "../models/policy";

const DCAService = {
    /**
     * Posts a new policy to the API.
     * @param {Policy} policy - The policy to be created.
     * @returns {Promise<Object>} A promise that resolves to the created policy.
     */
    createPolicy: async (policy: Policy) => {
        try {
            const endpoint = '/plugin/policy';
            const newPolicy = await post(endpoint, policy);
            return newPolicy;
        } catch (error) {
            console.error('Error creating policy:', error);
            throw error;
        }
    },
};

export default DCAService;
