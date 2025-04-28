import { useState, useEffect } from "react";
import Button from "@/modules/core/components/ui/button/Button";
import { isSupportedChainType, derivePathMap, toHex } from "@/modules/shared/wallet/wallet.utils";
import VulticonnectWalletService, { IVulticonnectVault }  from "@/modules/shared/wallet/vulticonnectWalletService";
import PricingService from "@/modules/policy/services/pricingService";

const getAccountVault = async (): Promise<[string, IVulticonnectVault]> => {
  const chain = localStorage.getItem("chain") as string;

  let accounts = [];
  if (!isSupportedChainType(chain)) {
    throw new Error('Chain not supported');
  }

  if (chain === "ethereum") {
    accounts = await VulticonnectWalletService.getConnectedEthAccounts();
  }
  if (!accounts || accounts.length === 0) {
    throw new Error("Need to connect to wallet");
  }

  const vaults = await VulticonnectWalletService.getVaults();

  const [account] = accounts;
  const [vault] = vaults;

  return [account, vault];
};

const sign = async (account: string, content: string): Promise<string> => {
  const hexMessage = toHex(content);

  const signature = await VulticonnectWalletService.signCustomMessage(
    hexMessage,
    account
  );

  return signature;
};

const ApproveFeeButton = ({
  pluginUrl,
  pluginType
}: {
  pluginUrl: string,
  pluginType: string
}) => {
  const [dcaIsApproved, setDcaIsApproved] = useState(false);

  useEffect(() => {
    const fetchDcaPluginPricingPolicy = async (): Promise<void> => {
      const pricing = await PricingService.getPluginPricing(pluginUrl, pluginType);
      setDcaIsApproved(pricing ? true : false);
    };

    fetchDcaPluginPricingPolicy();
  }, [pluginUrl, pluginType]);

  const handleApproveFee = async () => {
    if (pluginType !== "dca") {
      console.error("Only dca fee approvals are supported for now");
      return;
    }

    const pricingPolicy = {
      type: "PER_TX",
      amount: 0.1,
      metric: "PERCENTAGE"
    };

    const [account, vault] = await getAccountVault();
    const signature = await sign(account, JSON.stringify(pricingPolicy));

    const pricing = await PricingService.createPricing(pluginUrl, {
      public_key_ecdsa: vault.publicKeyEcdsa,
      public_key_eddsa: vault.publicKeyEddsa,
      plugin_type: "dca",
      is_ecdsa: true,
      chain_code_hex: vault.hexChainCode,
      derive_path: derivePathMap.ethereum,
      signature,
      policy: pricingPolicy
    });

    if (pricing.id) {
      console.log('plugin pricing created', pricing);
      setDcaIsApproved(true);
    }
  };

  return (
    <>
      {!dcaIsApproved && <Button
        size="small"
        type="button"
        styleType="primary"
        onClick={handleApproveFee}
      >
        Approve Fee
      </Button>}
    </>
  );
};

export default ApproveFeeButton;
