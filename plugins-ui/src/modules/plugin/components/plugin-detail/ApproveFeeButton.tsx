import { useState, useEffect } from "react";
import Button from "@/modules/core/components/ui/button/Button";
import { isSupportedChainType } from "@/modules/shared/wallet/wallet.utils";
import VulticonnectWalletService, { IVulticonnectVault }  from "@/modules/shared/wallet/vulticonnectWalletService";
import PricingService from '@/modules/policy/services/pricingService';

const ApproveFeeButton = ({
  pluginUrl,
  pluginType
}: {
  pluginUrl: string,
  pluginType: string
}) => {
  const [dcaIsApproved, setDcaIsApproved] = useState(false);

  useEffect(() => {
    fetchDcaPluginPricingPolicy();
  }, []);

  const fetchDcaPluginPricingPolicy = async (): Promise<void> => {
    const pricing = await PricingService.getPluginPricing(pluginUrl, pluginType);
    console.log('plugin pricing set', pricing);
    setDcaIsApproved(pricing ? true : false);
  };

  const handleApproveFee = () => {
    // TODO
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
  )
};

export default ApproveFeeButton;
