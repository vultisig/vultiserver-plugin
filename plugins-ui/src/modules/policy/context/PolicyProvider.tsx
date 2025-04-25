import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { useParams } from "react-router-dom";
import {
  PluginPolicy,
  PolicySchema,
  TransactionHistory,
} from "../models/policy";
import PolicyService from "../services/policyService";
import {
  derivePathMap,
  isSupportedChainType,
  toHex,
} from "@/modules/shared/wallet/wallet.utils";
import Toast from "@/modules/core/components/ui/toast/Toast";
import VulticonnectWalletService from "@/modules/shared/wallet/vulticonnectWalletService";
import { isEcdsaChain } from "@/modules/policy/utils/policy.util";
import MarketplaceService from "@/modules/marketplace/services/marketplaceService";
import { sortObjectAlphabetically } from "../utils/policy.util";

export const POLICY_ITEMS_PER_PAGE = 15;

export interface PolicyContextType {
  pluginType: string;
  policyMap: Map<string, PluginPolicy>;
  policySchemaMap: Map<string, PolicySchema>;
  policiesTotalCount: number;
  addPolicy: (policy: PluginPolicy) => Promise<boolean>;
  updatePolicy: (policy: PluginPolicy) => Promise<boolean>;
  removePolicy: (policyId: string) => Promise<void>;
  getPolicyHistory: (
    policyId: string,
    skip: number,
    take: number
  ) => Promise<TransactionHistory | null>;
  currentPage: number;
  setCurrentPage: (page: number) => void;
}

export const PolicyContext = createContext<PolicyContextType | undefined>(
  undefined
);

export const PolicyProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [policyMap, setPolicyMap] = useState(new Map<string, PluginPolicy>());
  const [currentPage, setCurrentPage] = useState(0);
  const [policiesTotalCount, setPoliciesTotalCount] = useState(0);
  const [policySchemaMap, setPolicySchemaMap] = useState(
    new Map<string, any>()
  );
  const [toast, setToast] = useState<{
    message: string;
    error?: string;
    type: "success" | "error";
  } | null>(null);

  const { pluginId } = useParams();
  const [pluginType, setPluginType] = useState("");
  const [serverEndpoint, setServerEndpoint] = useState("");
  const [authToken, setAuthToken] = useState(
    localStorage.getItem("authToken") || ""
  );

  const fetchPolicies = useCallback(async (): Promise<void> => {
    if (pluginType) {
      const fetchedPolicies = await MarketplaceService.getPolicies(
        pluginType,
        currentPage > 1 ? (currentPage - 1) * POLICY_ITEMS_PER_PAGE : 0,
        POLICY_ITEMS_PER_PAGE
      );

      const constructPolicyMap: Map<string, PluginPolicy> = new Map(
        fetchedPolicies?.policies?.map((p: PluginPolicy) => [p.id, p]) // Convert the array into [key, value] pairs
      );

      setPoliciesTotalCount(fetchedPolicies.total_count);
      setPolicyMap(constructPolicyMap);
    }
  }, [pluginType, currentPage]);

  useEffect(() => {
    const handleStorageChange = () => {
      setAuthToken(localStorage.getItem("authToken") || "");
    };

    // Listen for storage changes
    window.addEventListener("storage", handleStorageChange);

    const fetchPlugin = async (): Promise<void> => {
      if (!pluginId) return;

      try {
        const fetchedPlugin = await MarketplaceService.getPlugin(pluginId);

        if (fetchedPlugin) {
          setPluginType(fetchedPlugin.type);
          setServerEndpoint(fetchedPlugin.server_endpoint);

          const fetchPolicySchema = async (
            pluginType: string
          ): Promise<any> => {
            if (policySchemaMap.has(pluginType)) {
              return Promise.resolve(policySchemaMap.get(pluginType));
            }

            try {
              const fetchedSchemas = await PolicyService.getPolicySchema(
                fetchedPlugin.server_endpoint,
                fetchedPlugin.type
              );

              setPolicySchemaMap((prev) =>
                new Map(prev).set(fetchedPlugin.type, fetchedSchemas)
              );

              return Promise.resolve(fetchedSchemas);
            } catch (error: any) {
              console.error("Failed to fetch policy schema:", error.message);
              setToast({
                message: error.message || "Failed to fetch policy schema",
                error: error.error,
                type: "error",
              });

              return Promise.resolve(null);
            }
          };

          fetchPolicySchema(fetchedPlugin.type);
        }
      } catch (error: any) {
        console.error("Plugin not found:", error.message);
        setToast({
          message: "Plugin not found",
          error: error.error,
          type: "error",
        });

        return;
      }
    };

    fetchPlugin();

    return () => {
      window.removeEventListener("storage", handleStorageChange);
    };
  }, [authToken]);

  useEffect(() => {
    fetchPolicies().catch((error: any) => {
      console.error("Failed to get policies:", error.message);
      setToast({
        message: error.message || "Failed to get policies",
        error: error.error,
        type: "error",
      });
    });
  }, [fetchPolicies]);

  const addPolicy = async (policy: PluginPolicy): Promise<boolean> => {
    try {
      policy = await signPolicy(policy);
      if (policy.signature) {
        const newPolicy = await PolicyService.createPolicy(
          serverEndpoint,
          policy
        );
        setPolicyMap((prev) => new Map(prev).set(newPolicy.id, newPolicy));
        setToast({ message: "Policy created successfully!", type: "success" });

        return Promise.resolve(true);
      }
      return Promise.resolve(false);
    } catch (error: any) {
      console.error("Failed to create policy:", error.message);
      setToast({
        message: error.message || "Failed to create policy",
        error: error.error,
        type: "error",
      });

      return Promise.resolve(false);
    }
  };

  const updatePolicy = async (policy: PluginPolicy): Promise<boolean> => {
    try {
      policy = await signPolicy(policy);

      if (policy.signature) {
        const updatedPolicy = await PolicyService.updatePolicy(
          serverEndpoint,
          policy
        );

        setPolicyMap((prev) =>
          new Map(prev).set(updatedPolicy.id, updatedPolicy)
        );
        setToast({ message: "Policy updated successfully!", type: "success" });

        return Promise.resolve(true);
      }

      return Promise.resolve(false);
    } catch (error: any) {
      console.error("Failed to update policy:", error.message, error);
      setToast({
        message: error.message || "Failed to update policy",
        error: error.error,
        type: "error",
      });

      return Promise.resolve(false);
    }
  };

  const removePolicy = async (policyId: string): Promise<void> => {
    let policy = policyMap.get(policyId);

    if (!policy) return;

    try {
      policy = await signPolicy(policy);
      if (policy.signature) {
        await PolicyService.deletePolicy(serverEndpoint, policyId, policy.signature);

        setPolicyMap((prev) => {
          const updatedPolicyMap = new Map(prev);
          updatedPolicyMap.delete(policyId);

          return updatedPolicyMap;
        });

        setToast({
          message: "Policy deleted successfully!",
          type: "success",
        });
      }
    } catch (error: any) {
      console.error("Failed to delete policy:", error);
      setToast({
        message: error.message || "Failed to delete policy",
        error: error.error,
        type: "error",
      });
    }
  };

  const signPolicy = async (policy: PluginPolicy): Promise<PluginPolicy> => {
    const chain = localStorage.getItem("chain") as string;
    if (!isSupportedChainType(chain)) {
      // fail silently
      return policy
    }

    let accounts = [];
    if (chain === "ethereum") {
      accounts = await VulticonnectWalletService.getConnectedEthAccounts();
    }

    if (!accounts || accounts.length === 0) {
      throw new Error("Need to connect to wallet");
    }

    const vaults = await VulticonnectWalletService.getVaults();
    const [vault] = vaults

    // TODO: Only Ethereum currently supported
    const chainId = policy.policy.chain_id as string

    policy.public_key_ecdsa = vault.publicKeyEcdsa;
    policy.public_key_eddsa = vault.publicKeyEddsa;
    policy.is_ecdsa = isEcdsaChain(chainId);
    policy.chain_code_hex = vault.hexChainCode;
    policy.derive_path = derivePathMap[chain];

    const excludedFromSignature = {
      signature: "",
      progress: ""
    };
    const signPolicy = Object.assign({}, policy, excludedFromSignature);

    const policyWithSortedProperties = sortObjectAlphabetically(signPolicy);
    const serializedPolicy = JSON.stringify(policyWithSortedProperties);
    const hexMessage = toHex(serializedPolicy);

    const signature = await VulticonnectWalletService.signCustomMessage(
      hexMessage,
      accounts[0]
    );

    policy.signature = signature;

    console.log("Public key ecdsa: ", policy.public_key_ecdsa);
    console.log("Public key eddsa: ", policy.public_key_eddsa);
    console.log("Chain code hex: ", policy.chain_code_hex);
    console.log("Derive path: ", policy.derive_path);
    console.log("Hex message: ", hexMessage);
    console.log("Account[0]: ", accounts[0]);
    console.log("Signature: ", signature);

    return policy;
  };

  const getPolicyHistory = async (
    policyId: string,
    skip: number,
    take: number
  ): Promise<TransactionHistory | null> => {
    try {
      const history = await MarketplaceService.getPolicyTransactionHistory(
        policyId,
        skip,
        take
      );
      return history;
    } catch (error: any) {
      console.error("Failed to get policy history:", error);
      setToast({
        message: error.message,
        error: error.error,
        type: "error",
      });

      return null;
    }
  };

  return (
    <PolicyContext.Provider
      value={{
        pluginType,
        policyMap,
        policySchemaMap,
        policiesTotalCount,
        addPolicy,
        updatePolicy,
        removePolicy,
        getPolicyHistory,
        currentPage,
        setCurrentPage,
      }}
    >
      {children}
      {toast && (
        <Toast
          title={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}
    </PolicyContext.Provider>
  );
};

export const usePolicies = (): PolicyContextType => {
  const context = useContext(PolicyContext);
  if (!context) {
    throw new Error("usePolicies must be used within a PolicyProvider");
  }
  return context;
};
