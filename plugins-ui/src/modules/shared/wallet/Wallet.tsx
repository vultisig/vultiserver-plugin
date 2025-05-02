import Button from "@/modules/core/components/ui/button/Button";
import VulticonnectWalletService from "./vulticonnectWalletService";
import { useEffect, useState } from "react";
import {
  derivePathMap,
  getHexMessage,
  setLocalStorageAuthToken,
} from "./wallet.utils";
import MarketplaceService from "@/modules/marketplace/services/marketplaceService";
import { publish } from "@/utils/eventBus";

const Wallet = () => {
  let chain = localStorage.getItem("chain") as string;

  if (!chain) {
    localStorage.setItem("chain", "ethereum");
    chain = localStorage.getItem("chain") as string;
  }
  const [authToken, setAuthToken] = useState(
    localStorage.getItem("authToken") || ""
  );

  const [connectedWallet, setConnectedWallet] = useState(!!authToken);

  const connectWallet = async (chain: string) => {
    switch (chain) {
      // add more switch cases as more chains are supported
      case "ethereum": {
        try {
          const accounts =
            await VulticonnectWalletService.connectToVultiConnect();

          const vaults = await VulticonnectWalletService.getVaults();

          const publicKey = vaults[0].publicKeyEcdsa;
          if (publicKey) {
            localStorage.setItem("publicKey", publicKey);
          }
          const chainCodeHex = vaults[0].hexChainCode;
          const derivePath = derivePathMap[chain];

          const hexMessage = getHexMessage(publicKey);

          const signature = await VulticonnectWalletService.signCustomMessage(
            hexMessage,
            accounts[0]
          );

          if (signature && typeof signature === "string") {
            const token = await MarketplaceService.getAuthToken(
              hexMessage,
              signature,
              publicKey,
              chainCodeHex,
              derivePath
            );
            setLocalStorageAuthToken(token);
            setAuthToken(token);
          }
        } catch (error) {
          if (error instanceof Error) {
            console.error("Failed to update policy:", error.message, error);
            publish("onToast", {
              message: "Wallet connection failed!",
              type: "error",
            });
          }
        }

        break;
      }

      default:
        publish("onToast", {
          message: `Chain ${chain} is currently not supported.`,
          type: "error",
        });
        break;
    }
  };

  useEffect(() => {
    const handleStorageChange = () => {
      const hasToken = !!localStorage.getItem("authToken");
      setConnectedWallet(hasToken);
    };

    // Listen for storage changes
    window.addEventListener("storage", handleStorageChange);

    return () => {
      window.removeEventListener("storage", handleStorageChange);
    };
  }, [authToken]);

  return (
    <Button
      size="medium"
      styleType="primary"
      type="button"
      onClick={() => connectWallet(chain)}
    >
      {connectedWallet ? "Connected" : "Connect Wallet"}
    </Button>
  );
};

export default Wallet;
