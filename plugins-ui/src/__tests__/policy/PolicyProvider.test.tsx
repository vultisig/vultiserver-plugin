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
import { mockEventBus } from "../utils/global-mocks";
import userEvent from "@testing-library/user-event";

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

const mockPlugin = {
  id: "1",
  type: "type",
  title: "Plugin title",
  description: "Plugin description",
  metadata: {},
  server_endpoint: "endpoint",
  pricing_id: "pricingId",
};

const hoisted = vi.hoisted(() => ({
  mockPolicyService: {
    createPolicy: vi.fn(),
    updatePolicy: vi.fn(),
    deletePolicy: vi.fn(),
    getPolicySchema: vi.fn(),
  },
}));

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = (await importOriginal()) as Record<string, unknown>;
  return {
    ...actual,
    useParams: vi.fn(),
  };
});

vi.mock("@/modules/marketplace/services/marketplaceService", () => ({
  default: {
    getPlugin: vi.fn(),
    getPolicies: vi.fn(),
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
  const { policyMap, addPolicy, updatePolicy, removePolicy } = usePolicies();

  return (
    <div>
      <ul>
        {[...policyMap.values()].map((policy) => (
          <li key={policy.id}>{policy.id}</li>
        ))}
      </ul>

      <button
        onClick={() =>
          addPolicy({
            id: "3",
            public_key_ecdsa: "public_key_1_ecdsa",
            public_key_eddsa: "public_key_1_eddsa",
            is_ecdsa: true,
            chain_code_hex: "",
            derive_path: "",
            plugin_version: "0.0.1",
            policy_version: "0.0.1",
            plugin_type: "plugin_type",
            active: true,
            signature: "signature",
            policy: {},
            progress: PluginProgress.InProgress,
          })
        }
      >
        Add Policy
      </button>

      <button
        onClick={() =>
          updatePolicy({
            id: "2",
            public_key_ecdsa: "public_key_1_ecdsa",
            public_key_eddsa: "public_key_1_eddsa",
            is_ecdsa: true,
            chain_code_hex: "",
            derive_path: "",
            plugin_version: "0.0.1",
            policy_version: "0.0.1",
            plugin_type: "plugin_type",
            active: true,
            signature: "signature",
            policy: {},
            progress: PluginProgress.InProgress,
          })
        }
      >
        Update Policy
      </button>

      <button onClick={() => removePolicy("2")}>Delete Policy</button>
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

  describe("getPolicies", () => {
    it("should fetch & store policies in context", async () => {
      (useParams as Mock).mockReturnValue({ pluginId: "1" });

      (MarketplaceService.getPlugin as Mock).mockResolvedValue(mockPlugin);
      (MarketplaceService.getPolicies as Mock).mockResolvedValue(mockPolicies);
      renderWithProvider();

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
      });
    });

    it("should handle API failure and set toast error when getPolicies request fails", async () => {
      const mockError = new Error("API Error");

      (useParams as Mock).mockReturnValue({ pluginId: "1" });

      (MarketplaceService.getPlugin as Mock).mockResolvedValue(mockPlugin);
      (MarketplaceService.getPolicies as Mock).mockRejectedValue(mockError);

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
      (useParams as Mock).mockReturnValue({ pluginId: "1" });
      (MarketplaceService.getPlugin as Mock).mockResolvedValue(mockPlugin);
      (MarketplaceService.getPolicies as Mock).mockResolvedValue(mockPolicies);

      (PolicyService.createPolicy as Mock).mockResolvedValue({
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
      (useParams as Mock).mockReturnValue({ pluginId: "1" });

      (MarketplaceService.getPlugin as Mock).mockResolvedValue(mockPlugin);
      (MarketplaceService.getPolicies as Mock).mockResolvedValue(mockPolicies);

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
      (useParams as Mock).mockReturnValue({ pluginId: "1" });

      (MarketplaceService.getPlugin as Mock).mockResolvedValue(mockPlugin);
      (MarketplaceService.getPolicies as Mock).mockResolvedValue(mockPolicies);

      (PolicyService.updatePolicy as Mock).mockResolvedValue({
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
      (useParams as Mock).mockReturnValue({ pluginId: "1" });

      (MarketplaceService.getPlugin as Mock).mockResolvedValue(mockPlugin);
      (MarketplaceService.getPolicies as Mock).mockResolvedValue(mockPolicies);

      (PolicyService.updatePolicy as Mock).mockRejectedValue(
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
});
