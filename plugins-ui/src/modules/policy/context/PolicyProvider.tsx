import React, { createContext, useContext, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import {
  PluginProgress,
  PluginPolicy,
  PolicySchema,
  PolicyTransactionHistory,
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

export interface PolicyContextType {
  pluginType: string;
  policyMap: Map<string, PluginPolicy>;
  policySchemaMap: Map<string, PolicySchema>;
  addPolicy: (policy: PluginPolicy) => Promise<boolean>;
  updatePolicy: (policy: PluginPolicy) => Promise<boolean>;
  removePolicy: (policyId: string) => Promise<void>;
  getPolicyHistory: (policyId: string) => Promise<PolicyTransactionHistory[]>;
}

export const PolicyContext = createContext<PolicyContextType | undefined>(
  undefined
);

export const PolicyProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [policyMap, setPolicyMap] = useState(new Map<string, PluginPolicy>());
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
          const fetchPolicies = async (): Promise<void> => {
            try {
              const fetchedPolicies = await MarketplaceService.getPolicies(
                fetchedPlugin.type
              );

              const constructPolicyMap: Map<string, PluginPolicy> = new Map(
                fetchedPolicies?.map((p: PluginPolicy) => [p.id, p]) // Convert the array into [key, value] pairs
              );

              setPolicyMap(constructPolicyMap);
            } catch (error: any) {
              console.error("Failed to get policies:", error.message);
              setToast({
                message: error.message || "Failed to get policies",
                error: error.error,
                type: "error",
              });
            }
          };

          fetchPolicies();

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

  const addPolicy = async (policy: PluginPolicy): Promise<boolean> => {
    try {
      const signature = await signPolicy(policy);
      if (signature && typeof signature === "string") {
        policy.signature = signature;
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
      const signature = await signPolicy(policy);

      if (signature && typeof signature === "string") {
        policy.signature = signature;
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
    const policy = policyMap.get(policyId);

    if (!policy) return;

    try {
      const signature = await signPolicy(policy);
      if (signature && typeof signature === "string") {
        await PolicyService.deletePolicy(serverEndpoint, policyId, signature);

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

  const signPolicy = async (policy: PluginPolicy): Promise<string> => {
    const chain = localStorage.getItem("chain") as string;

    if (isSupportedChainType(chain)) {
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
      policy.signature = "";
      policy.is_ecdsa = isEcdsaChain(chainId);
      policy.chain_code_hex = vault.hexChainCode;
      policy.derive_path = derivePathMap[chain];
      const policyWithSortedProperties = sortObjectAlphabetically(policy);

      const serializedPolicy = JSON.stringify(policyWithSortedProperties);
      const hexMessage = toHex(serializedPolicy);

      const signature = await VulticonnectWalletService.signCustomMessage(
        hexMessage,
        accounts[0]
      );

      console.log("Public key ecdsa: ", policy.public_key_ecdsa);
      console.log("Public key eddsa: ", policy.public_key_eddsa);
      console.log("Chain code hex: ", policy.chain_code_hex);
      console.log("Derive path: ", policy.derive_path);
      console.log("Hex message: ", hexMessage);
      console.log("Account[0]: ", accounts[0]);
      console.log("Signature: ", signature);

      return signature;
    }
    return "";
  };

  const getPolicyHistory = async (
    policyId: string
  ): Promise<PolicyTransactionHistory[]> => {
    try {
      const history =
        await MarketplaceService.getPolicyTransactionHistory(policyId);
      return history;
    } catch (error: any) {
      console.error("Failed to get policy history:", error);
      setToast({
        message: error.message,
        error: error.error,
        type: "error",
      });

      return [];
    }
  };

  return (
    <PolicyContext.Provider
      value={{
        pluginType,
        policyMap,
        policySchemaMap,
        addPolicy,
        updatePolicy,
        removePolicy,
        getPolicyHistory,
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
