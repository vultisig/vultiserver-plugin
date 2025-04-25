import { get, post, put, remove } from "@/modules/core/services/httpService";
import MarketplaceService from "@/modules/marketplace/services/marketplaceService";
import {
  PluginPolicy,
  PolicySchema,
  PolicyTransactionHistory,
} from "@/modules/policy/models/policy";
import PolicyService from "@/modules/policy/services/policyService";
import { generatePolicy } from "@/modules/policy/utils/policy.util";
import { describe, it, expect, vi, afterEach, Mock, beforeEach } from "vitest";

vi.mock("@/modules/core/services/httpService", () => ({
  post: vi.fn(),
  put: vi.fn(),
  get: vi.fn(),
  remove: vi.fn(),
}));

describe("PolicyService", () => {
  beforeEach(() => {
    vi.stubEnv("VITE_MARKETPLACE_URL", "https://mock-api.com");
    localStorage.setItem("publicKey", "publicKey");
  });
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllEnvs();
    localStorage.clear();
  });

  describe("createPolicy", () => {
    it("should call /plugin/policy endpoint and return json object", async () => {
      const mockPolicy: PluginPolicy = generatePolicy(
        "",
        "",
        "pluginType",
        "",
        {}
      );
      const mockResponse: PluginPolicy = {
        id: "1",
        public_key_ecdsa: "public_key_ecdsa",
        public_key_eddsa: "public_key_eddsa",
        plugin_type: "pluginType",
        is_ecdsa: true,
        chain_code_hex: "",
        derive_path: "",
        plugin_version: "0.0.1",
        policy_version: "0.0.1",
        active: true,
        signature: "signature",
        policy: {},
        progress: "IN PROGRESS",
      };

      (post as Mock).mockResolvedValue(mockResponse);

      const result = await PolicyService.createPolicy(
        "https://mock-api.com",
        mockPolicy
      );

      expect(post).toHaveBeenCalledWith(
        "https://mock-api.com/plugin/policy",
        mockPolicy
      );
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when post fails", async () => {
      const mockPolicy: PluginPolicy = generatePolicy(
        "",
        "",
        "pluginType",
        "",
        {}
      );
      const mockError = new Error("API Error");

      (post as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      await expect(
        PolicyService.createPolicy("https://mock-api.com", mockPolicy)
      ).rejects.toThrow("API Error");

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error creating policy:",
        mockError
      );
    });
  });

  describe("updatePolicy", () => {
    it("should call /plugin/policy endpoint and return json object", async () => {
      const mockPolicy: PluginPolicy = generatePolicy(
        "",
        "",
        "pluginType",
        "",
        {}
      );
      const mockResponse: PluginPolicy = {
        id: "1",
        public_key_ecdsa: "public_key_ecdsa",
        public_key_eddsa: "public_key_eddsa",
        plugin_type: "pluginType",
        is_ecdsa: true,
        chain_code_hex: "",
        derive_path: "",
        plugin_version: "0.0.1",
        policy_version: "0.0.1",
        active: true,
        signature: "signature",
        policy: {},
        progress: "IN PROGRESS",
      };

      (put as Mock).mockResolvedValue(mockResponse);

      const result = await PolicyService.updatePolicy(
        "https://mock-api.com",
        mockPolicy
      );

      expect(put).toHaveBeenCalledWith(
        "https://mock-api.com/plugin/policy",
        mockPolicy
      );
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when put fails", async () => {
      const mockPolicy: PluginPolicy = generatePolicy(
        "",
        "",
        "pluginType",
        "",
        {}
      );
      const mockError = new Error("API Error");

      (put as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      await expect(
        PolicyService.updatePolicy("https://mock-api.com", mockPolicy)
      ).rejects.toThrow("API Error");

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error updating policy:",
        mockError
      );
    });
  });

  describe("getPolicies", () => {
    it("should call /plugins/policies endpoint and return json object", async () => {
      const PUBLIC_KEY = localStorage.getItem("publicKey");
      const mockRequest = {
        headers: {
          Authorization: "Bearer null",
          plugin_type: "pluginType",
          public_key: PUBLIC_KEY,
        },
      };
      const mockResponse: PluginPolicy[] = [
        {
          id: "1",
          public_key_ecdsa: "public_key_ecdsa",
          public_key_eddsa: "public_key_eddsa",
          plugin_type: "pluginType",
          active: true,
          is_ecdsa: true,
          chain_code_hex: "",
          derive_path: "",
          plugin_version: "0.0.1",
          policy_version: "0.0.1",
          signature: "signature",
          policy: {},
          progress: "IN PROGRESS",
        },
      ];

      (get as Mock).mockResolvedValue(mockResponse);

      const result = await MarketplaceService.getPolicies("pluginType", 0, 10);

      expect(get).toHaveBeenCalledWith(
        "https://mock-api.com/plugins/policies?skip=0&take=10",
        mockRequest
      );
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when get fails", async () => {
      const mockError = new Error("API Error");

      (get as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      await expect(
        MarketplaceService.getPolicies("pluginType", 0, 10)
      ).rejects.toThrow("API Error");

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error getting policies:",
        mockError
      );
    });
  });

  describe("getPolicyTransactionHistory", () => {
    it("should call /plugins/policies/{policyId}/history endpoint and return json object", async () => {
      const PUBLIC_KEY = localStorage.getItem("publicKey");

      const mockRequest = {
        headers: {
          Authorization: "Bearer null",
          public_key: PUBLIC_KEY,
        },
      };
      const mockResponse: PolicyTransactionHistory[] = [
        {
          id: "1",
          updated_at: "03/07/25",
          status: "MINED",
        },
      ];

      (get as Mock).mockResolvedValue(mockResponse);

      const result = await MarketplaceService.getPolicyTransactionHistory(
        "policyId",
        0,
        10
      );

      expect(get).toHaveBeenCalledWith(
        "https://mock-api.com/plugins/policies/policyId/history?skip=0&take=10",
        mockRequest
      );
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when get fails", async () => {
      const mockError = new Error("API Error");

      (get as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      await expect(
        MarketplaceService.getPolicyTransactionHistory("policyId", 0, 10)
      ).rejects.toThrow("API Error");

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error getting policy history:",
        mockError
      );
    });
  });

  describe("deletePolicy", () => {
    it("should call /plugin/policy/{policyId} endpoint and return nothing", async () => {
      (remove as Mock).mockResolvedValue(undefined);

      const result = await PolicyService.deletePolicy(
        "https://mock-api.com",
        "policyId",
        "signature"
      );

      expect(remove).toHaveBeenCalledWith(
        "https://mock-api.com/plugin/policy/policyId",
        {
          signature: "signature",
        }
      );
      expect(result).toEqual(undefined);
    });

    it("throws an error when remove fails", async () => {
      const mockError = new Error("API Error");

      (remove as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      await expect(
        PolicyService.deletePolicy(
          "https://mock-api.com",
          "policyId",
          "signature"
        )
      ).rejects.toThrow("API Error");

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error deleting policy:",
        mockError
      );
    });
  });

  describe("getPolicySchema", () => {
    it("should call /plugins/schema endpoint and return json object", async () => {
      const mockRequest = {
        headers: {
          plugin_type: "pluginType",
        },
      };
      const mockResponse: PolicySchema[] = [
        {
          form: {
            schema: {},
            uiSchema: {},
            plugin_version: "",
            policy_version: "",
            plugin_type: "",
          },
          table: {
            columns: [],
            mapping: {},
          },
        },
      ];

      (get as Mock).mockResolvedValue(mockResponse);

      const result = await PolicyService.getPolicySchema(
        "https://mock-api.com",
        "pluginType"
      );

      expect(get).toHaveBeenCalledWith(
        "https://mock-api.com/plugin/policy/schema",
        mockRequest
      );
      expect(result).toEqual(mockResponse);
    });

    it("throws an error when get fails", async () => {
      const mockError = new Error("API Error");

      (get as Mock).mockRejectedValue(mockError);
      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      await expect(
        PolicyService.getPolicySchema("https://mock-api.com", "pluginType")
      ).rejects.toThrow("API Error");

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        "Error getting policy schema:",
        mockError
      );
    });
  });
});
