import { render, screen, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, Mock, afterEach } from "vitest";
import PolicyService from "@/modules/policy/services/policyService";
import {
  PolicyProvider,
  usePolicies,
} from "@/modules/policy/context/PolicyProvider";
import VulticonnectWalletService from "@/modules/shared/wallet/vulticonnectWalletService";
import MarketplaceService from "@/modules/marketplace/services/marketplaceService";
import {
  PluginProgress,
  PluginPoliciesMap,
} from "@/modules/policy/models/policy";
import { useParams } from "react-router-dom";
import {
  mockEventBus,
  mockPlugin,
  mockPluginPolicy,
} from "@tests-utils/global-mocks";
import userEvent from "@testing-library/user-event";
import { act } from "react";

const mockPolicies: PluginPoliciesMap = {
  policies: [
    {
      id: "1",
      public_key_ecdsa: "public_key_1_ecdsa",
      public_key_eddsa: "public_key_1_eddsa",
      plugin_type: "plugin_type",
      active: true,
      signature: "signature",
      policy: {},
      is_ecdsa: true,
      chain_code_hex: "chain_code_hex",
      derive_path: "derive_path",
      progress: PluginProgress.InProgress,
      plugin_version: "0.01",
      policy_version: "0.01",
    },
    {
      id: "2",
      public_key_ecdsa: "public_key_2_ecdsa",
      public_key_eddsa: "public_key_2_eddsa",
      plugin_type: "plugin_type",
      active: false,
      signature: "signature",
      policy: {},
      is_ecdsa: true,
      chain_code_hex: "chain_code_hex",
      derive_path: "derive_path",
      progress: PluginProgress.InProgress,
      plugin_version: "0.01",
      policy_version: "0.01",
    },
  ],
  total_count: 2,
};

const hoisted = vi.hoisted(() => ({
  mockPolicyService: {
    createPolicy: vi.fn(),
    updatePolicy: vi.fn(),
    deletePolicy: vi.fn(),
    getPolicySchema: vi.fn(),
  },
  mockMarketplaceService: {
    getPlugin: vi.fn(() => mockPlugin),
    getPolicies: vi.fn(() => mockPolicies),
    getPolicyTransactionHistory: vi.fn(() => []),
  },
}));

vi.mock("react-router-dom", async (importOriginal) => {
  return {
    ...(await importOriginal()),
    useParams: vi.fn(() => ({ pluginId: "1" })),
  };
});

vi.mock("@/modules/marketplace/services/marketplaceService", () => ({
  default: {
    getPlugin: hoisted.mockMarketplaceService.getPlugin,
    getPolicies: hoisted.mockMarketplaceService.getPolicies,
    getPolicyTransactionHistory:
      hoisted.mockMarketplaceService.getPolicyTransactionHistory,
  },
}));

vi.mock("@/modules/policy/services/policyService", () => ({
  default: {
    createPolicy: hoisted.mockPolicyService.createPolicy,
    updatePolicy: hoisted.mockPolicyService.updatePolicy,
    deletePolicy: hoisted.mockPolicyService.deletePolicy,
    getPolicySchema: hoisted.mockPolicyService.getPolicySchema,
  },
}));

const TestComponent = () => {
  const { policyMap, addPolicy, updatePolicy, removePolicy, getPolicyHistory } =
    usePolicies();

  return (
    <div>
      <ul>
        {[...policyMap.values()].map((policy) => (
          <li key={policy.id}>{policy.id}</li>
        ))}
      </ul>

      <button onClick={() => addPolicy(mockPluginPolicy)}>Add Policy</button>

      <button onClick={() => updatePolicy(mockPluginPolicy)}>
        Update Policy
      </button>

      <button onClick={() => removePolicy("2")}>Delete Policy</button>
      <button onClick={() => getPolicyHistory("2", 1, 2)}>
        Get Policy History
      </button>
    </div>
  );
};

const renderWithProvider = () => {
  return render(
    <PolicyProvider>
      <TestComponent />
    </PolicyProvider>
  );
};

describe("PolicyProvider", () => {
  beforeEach(() => {
    localStorage.setItem("chain", "ethereum");
    vi.spyOn(
      VulticonnectWalletService,
      "getConnectedEthAccounts"
    ).mockImplementation(() => Promise.resolve(["account address"]));

    vi.spyOn(VulticonnectWalletService, "signCustomMessage").mockImplementation(
      () => Promise.resolve("some hex signature")
    );

    window.vultisig = {
      getVaults: vi.fn().mockResolvedValue(["vault 1", "vault 2"]),
    };
  });

  afterEach(() => {
    vi.resetAllMocks();
    localStorage.clear();
  });

  describe("useEffect", () => {
    it("should call function in useEffect", async () => {
      act(() => {
        renderWithProvider();
      });
      await waitFor(() => {
        expect(hoisted.mockMarketplaceService.getPlugin).toBeCalledWith("1");
        expect(hoisted.mockMarketplaceService.getPolicies).toBeCalledWith(
          mockPlugin.type,
          0,
          15
        );
        expect(hoisted.mockPolicyService.getPolicySchema).toBeCalledWith(
          mockPlugin.server_endpoint,
          mockPlugin.type
        );
      });
    });
    it("should show toast message if getPlugin fails", async () => {
      hoisted.mockMarketplaceService.getPlugin.mockRejectedValueOnce(
        new Error("From tests")
      );
      act(() => {
        renderWithProvider();
      });
      await waitFor(() => {
        expect(hoisted.mockMarketplaceService.getPlugin).toBeCalledWith("1");
        expect(mockEventBus.publish).toBeCalledWith("onToast", {
          message: "Plugin not found",
          type: "error",
        });
      });
    });
    it("should show toast message if getPolicies fails", async () => {
      hoisted.mockMarketplaceService.getPolicies.mockRejectedValueOnce(
        new Error("From tests")
      );
      act(() => {
        renderWithProvider();
      });
      await waitFor(() => {
        expect(hoisted.mockMarketplaceService.getPlugin).toBeCalledWith("1");
        expect(mockEventBus.publish).toBeCalledWith("onToast", {
          message: "From tests",
          type: "error",
        });
      });
    });
    it("should show toast message if getPolicySchema fails", async () => {
      hoisted.mockPolicyService.getPolicySchema.mockRejectedValueOnce(
        new Error("From tests")
      );
      act(() => {
        renderWithProvider();
      });
      await waitFor(() => {
        expect(mockEventBus.publish).toBeCalledWith("onToast", {
          message: "From tests",
          type: "error",
        });
      });
    });
  });

  describe("getPolicies", () => {
    it("should fetch & store policies in context", async () => {
      renderWithProvider();

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
      });
    });

    it("should handle API failure and set toast error when getPolicies request fails", async () => {
      const mockError = new Error("API Error");
      hoisted.mockMarketplaceService.getPolicies.mockRejectedValue(mockError);

      renderWithProvider();

      await waitFor(() => {
        expect(mockEventBus.publish).toBeCalledWith("onToast", {
          type: "error",
          message: "API Error",
        });
      });
    });
  });

  describe("addPolicy", () => {
    it("should add policy in context", async () => {
      hoisted.mockPolicyService.createPolicy.mockResolvedValue({
        id: "3",
        public_key: "public_key_1",
        plugin_type: "plugin_type",
        active: true,
        signature: "signature",
        policy: {},
        is_ecdsa: true,
        chain_code_hex: "chain_code_hex",
        derive_path: "derive_path",
        plugin_id: "plugin_id",
      });

      renderWithProvider();

      const newPolicyButton = screen.getByRole("button", {
        name: "Add Policy",
      });

      await userEvent.click(newPolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(screen.getByText("3")).toBeInTheDocument();
        expect(mockEventBus.publish).toBeCalledWith("onToast", {
          message: "Policy created successfully!",
          type: "success",
        });
      });
    });

    it("should set error message if request fails", async () => {
      hoisted.mockPolicyService.createPolicy.mockRejectedValue(
        new Error("API Error")
      );

      renderWithProvider();

      const newPolicyButton = screen.getByRole("button", {
        name: "Add Policy",
      });

      await userEvent.click(newPolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(mockEventBus.publish).toBeCalledWith("onToast", {
          type: "error",
          message: "API Error",
        });
      });
    });
  });

  describe("updatePolicy", () => {
    it("should update policy in context", async () => {
      hoisted.mockPolicyService.updatePolicy.mockResolvedValue({
        id: "2",
        public_key: "public_key_1",
        plugin_type: "plugin_type",
        active: false,
        signature: "signature",
        policy: {},
        is_ecdsa: true,
        chain_code_hex: "chain_code_hex",
        derive_path: "derive_path",
        plugin_id: "plugin_id",
      });

      renderWithProvider();

      const updatePolicyButton = screen.getByRole("button", {
        name: "Update Policy",
      });

      await userEvent.click(updatePolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(mockEventBus.publish).toBeCalledWith("onToast", {
          message: "Policy updated successfully!",
          type: "success",
        });
      });
    });

    it("should set error message if request fails", async () => {
      hoisted.mockPolicyService.updatePolicy.mockRejectedValue(
        new Error("API Error")
      );

      renderWithProvider();

      const updatePolicyButton = screen.getByRole("button", {
        name: "Update Policy",
      });

      userEvent.click(updatePolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(mockEventBus.publish).toBeCalledWith("onToast", {
          message: "API Error",
          type: "error",
        });
      });
    });
  });

  describe("removePolicy", () => {
    it("should delete policy from context", async () => {
      (useParams as Mock).mockReturnValue({ pluginId: "1" });

      (MarketplaceService.getPlugin as Mock).mockResolvedValue(mockPlugin);
      (MarketplaceService.getPolicies as Mock).mockResolvedValue(mockPolicies);

      (PolicyService.deletePolicy as Mock).mockResolvedValue({});

      await renderWithProvider();

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
      });

      const deletePolicyButton = screen.getByRole("button", {
        name: "Delete Policy",
      });

      userEvent.click(deletePolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.queryByText("2")).not.toBeInTheDocument();
        expect(mockEventBus.publish).toBeCalledWith("onToast", {
          message: "Policy deleted successfully!",
          type: "success",
        });
      });
    });

    it("should set error message if request fails", async () => {
      (useParams as Mock).mockReturnValue({ pluginId: "1" });

      (MarketplaceService.getPlugin as Mock).mockResolvedValue(mockPlugin);
      (MarketplaceService.getPolicies as Mock).mockResolvedValue(mockPolicies);

      (PolicyService.deletePolicy as Mock).mockRejectedValue(
        new Error("API Error")
      );

      renderWithProvider();

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
      });

      const deletePolicyButton = screen.getByRole("button", {
        name: "Delete Policy",
      });

      userEvent.click(deletePolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(mockEventBus.publish).toHaveBeenCalledWith("onToast", {
          message: "API Error",
          type: "error",
        });
      });
    });
  });

  describe("removePolicy", () => {
    it("should remove policy", async () => {
      const { findByText } = renderWithProvider();
      const deleteButton = await findByText("Delete Policy");
      await userEvent.click(deleteButton);
      expect(hoisted.mockPolicyService.deletePolicy).toBeCalledWith(
        mockPlugin.server_endpoint,
        "2",
        mockPluginPolicy.signature
      );
      expect(mockEventBus.publish).toBeCalledWith("onToast", {
        message: "Policy deleted successfully!",
        type: "success",
      });
    });
    it("should set error message if request fails", async () => {
      hoisted.mockPolicyService.deletePolicy.mockRejectedValueOnce(
        new Error("From tests")
      );
      const { findByText } = renderWithProvider();
      const deleteButton = await findByText("Delete Policy");
      await userEvent.click(deleteButton);
      expect(hoisted.mockPolicyService.deletePolicy).toBeCalledWith(
        mockPlugin.server_endpoint,
        "2",
        mockPluginPolicy.signature
      );
      expect(mockEventBus.publish).toBeCalledWith("onToast", {
        message: "From tests",
        type: "error",
      });
    });
  });

  describe("removePolicy", () => {
    it("should remove policy", async () => {
      const { findByText } = renderWithProvider();
      const getHistoryButton = await findByText("Get Policy History");
      await userEvent.click(getHistoryButton);
      expect(
        hoisted.mockMarketplaceService.getPolicyTransactionHistory
      ).toBeCalledWith("2", 1, 2);
    });
    it("should set error message if request fails", async () => {
      hoisted.mockMarketplaceService.getPolicyTransactionHistory.mockRejectedValueOnce(
        new Error("From tests")
      );
      const { findByText } = renderWithProvider();
      const getHistoryButton = await findByText("Get Policy History");
      await userEvent.click(getHistoryButton);
      expect(
        hoisted.mockMarketplaceService.getPolicyTransactionHistory
      ).toBeCalledWith("2", 1, 2);
      expect(mockEventBus.publish).toBeCalledWith("onToast", {
        message: "From tests",
        type: "error",
      });
    });
  });
});
