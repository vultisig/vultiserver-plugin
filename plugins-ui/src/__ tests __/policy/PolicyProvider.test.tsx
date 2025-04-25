import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, Mock, afterEach } from "vitest";
import PolicyService from "@/modules/policy/services/policyService";
import {
  PolicyProvider,
  usePolicies,
} from "@/modules/policy/context/PolicyProvider";
import VulticonnectWalletService from "@/modules/shared/wallet/vulticonnectWalletService";
import MarketplaceService from "@/modules/marketplace/services/marketplaceService";
import { useParams } from "react-router-dom";

const mockPolicies = [
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
    plugin_id: "plugin_id",
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
    plugin_id: "plugin_id",
  },
];

const mockPlugin = {
  id: "1",
  type: "type",
  title: "Plugin title",
  description: "Plugin description",
  metadata: {},
  server_endpoint: "endpoint",
  pricing_id: "pricingId",
};

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = (await importOriginal()) as {};
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
    createPolicy: vi.fn(),
    updatePolicy: vi.fn(),
    deletePolicy: vi.fn(),
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

    (window as any).vultisig = {
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

      const consoleErrorSpy = vi
        .spyOn(console, "error")
        .mockImplementation(() => {});

      renderWithProvider();

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalledWith(
          "Failed to get policies:",
          "API Error"
        );

        const closeToastButton = screen.getByRole("button", {
          name: "Close message",
        });

        expect(closeToastButton).toBeInTheDocument();

        const errorMessage = screen.getByText("API Error");
        expect(errorMessage).toBeInTheDocument();
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

      fireEvent.click(newPolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(screen.getByText("3")).toBeInTheDocument();
        expect(
          screen.getByText("Policy created successfully!")
        ).toBeInTheDocument();
      });
    });

    it("should set error message if request fails", async () => {
      (useParams as Mock).mockReturnValue({ pluginId: "1" });

      (MarketplaceService.getPlugin as Mock).mockResolvedValue(mockPlugin);
      (MarketplaceService.getPolicies as Mock).mockResolvedValue(mockPolicies);

      (PolicyService.createPolicy as Mock).mockRejectedValue("API Error");

      renderWithProvider();

      const newPolicyButton = screen.getByRole("button", {
        name: "Add Policy",
      });

      fireEvent.click(newPolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(screen.queryByText("3")).not.toBeInTheDocument();
        expect(screen.getByText("Failed to create policy")).toBeInTheDocument();
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

      fireEvent.click(updatePolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(
          screen.getByText("Policy updated successfully!")
        ).toBeInTheDocument();
      });
    });

    it("should set error message if request fails", async () => {
      (useParams as Mock).mockReturnValue({ pluginId: "1" });

      (MarketplaceService.getPlugin as Mock).mockResolvedValue(mockPlugin);
      (MarketplaceService.getPolicies as Mock).mockResolvedValue(mockPolicies);

      (PolicyService.updatePolicy as Mock).mockRejectedValue("API Error");

      renderWithProvider();

      const updatePolicyButton = screen.getByRole("button", {
        name: "Update Policy",
      });

      fireEvent.click(updatePolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(screen.getByText("Failed to update policy")).toBeInTheDocument();
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

      fireEvent.click(deletePolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.queryByText("2")).not.toBeInTheDocument();
        expect(
          screen.getByText("Policy deleted successfully!")
        ).toBeInTheDocument();
      });
    });

    it("should set error message if request fails", async () => {
      (useParams as Mock).mockReturnValue({ pluginId: "1" });

      (MarketplaceService.getPlugin as Mock).mockResolvedValue(mockPlugin);
      (MarketplaceService.getPolicies as Mock).mockResolvedValue(mockPolicies);

      (PolicyService.deletePolicy as Mock).mockRejectedValue("API Error");

      renderWithProvider();

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
      });

      const deletePolicyButton = screen.getByRole("button", {
        name: "Delete Policy",
      });

      fireEvent.click(deletePolicyButton);

      await waitFor(() => {
        expect(screen.getByText("1")).toBeInTheDocument();
        expect(screen.getByText("2")).toBeInTheDocument();
        expect(screen.getByText("Failed to delete policy")).toBeInTheDocument();
      });
    });
  });
});
