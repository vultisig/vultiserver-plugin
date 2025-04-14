import { PluginMap } from "@/modules/marketplace/models/marketplace";
import MarketplaceService from "@/modules/marketplace/services/marketplaceService";
import { Plugin } from "@/modules/plugin/models/plugin";
import {
  PluginPolicy,
  PolicyTransactionHistory,
} from "@/modules/policy/models/policy";
import { describe, it, expect, vi } from "vitest";

const hoisted = vi.hoisted(() => ({
  mockGet: vi.fn(),
  mockPost: vi.fn(),
}));

vi.mock("@/modules/core/services/httpService", () => ({
  get: hoisted.mockGet,
  post: hoisted.mockPost,
}));
const MARKETPLACE_URL = import.meta.env.VITE_MARKETPLACE_URL;
describe("MarketplaceService", () => {
  describe("getPlugins", () => {
    it("should return PluginsMap", async () => {
      const mockedResult = { plugins: [], total_count: 0 } as PluginMap;
      hoisted.mockGet.mockResolvedValueOnce(mockedResult);
      const result = await MarketplaceService.getPlugins(
        "",
        "",
        "",
        "asc",
        1,
        2
      );
      expect(result).toStrictEqual(mockedResult);
      expect(hoisted.mockGet).toBeCalledWith(
        `${MARKETPLACE_URL}/plugins?term=&category_id=&sort=&skip=1&take=2`
      );
    });
    it("should throw and error and write to console.error", async () => {
      hoisted.mockGet.mockRejectedValueOnce(new Error("From test"));
      const consoleSpy = vi.spyOn(console, "error");
      await expect(
        MarketplaceService.getPlugins("", "", "", "asc", 1, 2)
      ).rejects.toThrow("From test");
      expect(consoleSpy).toBeCalledWith(
        "Error getting plugins:",
        new Error("From test")
      );
    });
  });
  describe("getPlugin", () => {
    it("should return Plugin", async () => {
      const mockedResult = {
        id: "1",
        type: "type",
        title: "Plugin",
        description: "Description",
        metadata: {},
        server_endpoint: "server",
        pricing_id: "100",
      } as Plugin;
      hoisted.mockGet.mockResolvedValueOnce(mockedResult);
      const result = await MarketplaceService.getPlugin("1");
      expect(result).toStrictEqual(mockedResult);
      expect(hoisted.mockGet).toBeCalledWith(`${MARKETPLACE_URL}/plugins/${1}`);
    });
    it("should throw and error and write to console.error", async () => {
      hoisted.mockGet.mockRejectedValueOnce(new Error("From test"));
      const consoleSpy = vi.spyOn(console, "error");
      await expect(MarketplaceService.getPlugin("1")).rejects.toThrow(
        "From test"
      );
      expect(consoleSpy).toBeCalledWith(
        "Error getting plugin:",
        new Error("From test")
      );
    });
  });
  describe("getAuthToken", () => {
    const mockedPayload: Parameters<typeof MarketplaceService.getAuthToken> = [
      "Message",
      "Signature",
      "Public key",
      "Chain codex HEX",
      "Derive path",
    ];
    it("should return token", async () => {
      const mockedResult = "token";
      hoisted.mockPost.mockResolvedValueOnce({ token: mockedResult });
      const result = await MarketplaceService.getAuthToken(...mockedPayload);
      expect(hoisted.mockPost).toBeCalledWith(`${MARKETPLACE_URL}/auth`, {
        message: mockedPayload[0],
        signature: mockedPayload[1],
        public_key: mockedPayload[2],
        chain_code_hex: mockedPayload[3],
        derive_path: mockedPayload[4],
      });
      expect(result).toEqual(mockedResult);
    });
    it("should throw and error and write to console.error", async () => {
      hoisted.mockPost.mockRejectedValueOnce(new Error("From test"));
      const consoleSpy = vi.spyOn(console, "error");
      await expect(
        MarketplaceService.getAuthToken(...mockedPayload)
      ).rejects.toThrow("From test");
      expect(consoleSpy).toBeCalledWith(
        "Failed to get auth token",
        new Error("From test")
      );
    });
  });
  describe("getPolicies", () => {
    it("should return PluginPolicy", async () => {
      const mockedResult = {
        id: "Id",
        public_key: "Public key",
        is_ecdsa: false,
        chain_code_hex: "Chain code HEX",
        derive_path: "Derive path",
        plugin_version: "v0.0.1",
        policy_version: "v0.0.2",
        plugin_type: "Plugin type",
        signature: "Signature",
        policy: {},
        active: false,
      } as PluginPolicy;
      hoisted.mockGet.mockResolvedValueOnce(mockedResult);
      const result = await MarketplaceService.getPolicies("test", 1, 2);
      expect(result).toStrictEqual(mockedResult);
      expect(hoisted.mockGet).toBeCalledWith(
        `${MARKETPLACE_URL}/plugins/policies?skip=1&take=2`,
        {
          headers: {
            plugin_type: "test",
            public_key: "null",
            Authorization: `Bearer ${localStorage.getItem("authToken")}`,
          },
        }
      );
    });
    it("should throw and error and write to console.error when mesage is not Unauthorized", async () => {
      hoisted.mockGet.mockRejectedValueOnce(new Error("From test"));
      const consoleSpy = vi.spyOn(console, "error");
      await expect(
        MarketplaceService.getPolicies("test", 1, 2)
      ).rejects.toThrow("From test");
      expect(consoleSpy).toBeCalledWith(
        "Error getting policies:",
        new Error("From test")
      );
    });
    it("should throw and error and write to console.error when mesage is Unauthorized", async () => {
      hoisted.mockGet.mockRejectedValueOnce(new Error("Unauthorized"));
      const consoleSpy = vi.spyOn(console, "error");
      const localStorageSpy = vi.spyOn(Storage.prototype, "removeItem");
      const windowDispatchSpy = vi.spyOn(window, "dispatchEvent");
      await expect(
        MarketplaceService.getPolicies("test", 1, 2)
      ).rejects.toThrow("Unauthorized");
      expect(consoleSpy).toBeCalledWith(
        "Error getting policies:",
        new Error("Unauthorized")
      );
      expect(localStorageSpy).toBeCalledWith("authToken");
      expect(windowDispatchSpy).toBeCalled();
    });
  });
  describe("getPolicyTransactionHistory", () => {
    it("should return PolicyTransactionHistory", async () => {
      const mockedResult = [
        {
          id: "ID",
          updated_at: "Updated At",
          status: "Status",
        },
      ] as PolicyTransactionHistory[];
      hoisted.mockGet.mockResolvedValueOnce(mockedResult);
      const result = await MarketplaceService.getPolicyTransactionHistory(
        "test",
        1,
        2
      );
      expect(result).toStrictEqual(mockedResult);
      expect(hoisted.mockGet).toBeCalledWith(
        `${MARKETPLACE_URL}/plugins/policies/test/history?skip=1&take=2`,
        {
          headers: {
            public_key: "null",
            Authorization: `Bearer ${localStorage.getItem("authToken")}`,
          },
        }
      );
    });
    it("should throw and error and write to console.error", async () => {
      hoisted.mockGet.mockRejectedValueOnce(new Error("From test"));
      const consoleSpy = vi.spyOn(console, "error");
      await expect(
        MarketplaceService.getPolicyTransactionHistory("test", 1, 2)
      ).rejects.toThrow("From test");
      expect(consoleSpy).toBeCalledWith(
        "Error getting policy history:",
        new Error("From test")
      );
    });
  });
});
