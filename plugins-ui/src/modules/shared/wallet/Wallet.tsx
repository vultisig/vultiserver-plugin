import Button from "@/modules/core/components/ui/button/Button";
import VulticonnectWalletService from "./vulticonnectWalletService";
import { useEffect, useState } from "react";
import PolicyService from "@/modules/policy/services/policyService";
import { derivePathMap, getHexMessage } from "./wallet.utils";

const Wallet = () => {
  const [chain, setChain] = useState(() => {
    const savedChain = localStorage.getItem("chain");
    return savedChain || "ethereum";
  });
  
  const [isConnecting, setIsConnecting] = useState(false);
  const [connectedWallet, setConnectedWallet] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Check if wallet is already connected on component mount
  useEffect(() => {
    const checkConnection = async () => {
      try {
        if (window.vultisig?.ethereum) {
          console.log("window.vultisig?.ethereum", window.vultisig?.ethereum);
          console.log("window.vultisig", window.vultisig);
          console.log("VulticonnectWalletService", VulticonnectWalletService);
          const accounts = await VulticonnectWalletService.getConnectedEthAccounts();
          console.log("accounts", accounts);
          if (accounts && accounts.length > 0) {
            setConnectedWallet(true);
          }
        }
      } catch (err) {
        console.error("Failed to check wallet connection:", err);
      }
    };
    
    checkConnection();
  }, []);

  const connectWallet = async (chain: string) => {
    console.log("connectWallet", chain);

    setIsConnecting(true);
    setError(null);
    
    if (!window.vultisig?.ethereum) {
      setError("Vultisg extension not found. Please install the Vultisg Chrome extension.");
      setIsConnecting(false);
      return;
    }

    try {
      switch (chain) {
        // add more switch cases as more chains are supported
        case "ethereum": {
          console.log("ethereum connection in progress");
          const accounts = await VulticonnectWalletService.connectToVultiConnect();
          console.log("ethereum connection in progress :: accounts", accounts);
          
          if (!accounts || !accounts.length) {
            throw new Error("Failed to connect to wallet");
          }
          
          const walletAddress = accounts[0];
          
          const vaults = await VulticonnectWalletService.getVaults();
          if (!vaults || !vaults.length) {
            throw new Error("No vaults found");
          }
          
          const publicKey = vaults[0].publicKeyEcdsa;
          if (!publicKey) {
            throw new Error("No public key found");
          }
          
          localStorage.setItem("publicKey", publicKey);
          localStorage.setItem("walletAddress", walletAddress);
          localStorage.setItem("chain", chain);
          
          const chainCodeHex = vaults[0].hexChainCode;
          const derivePath = derivePathMap[chain];

          const hexMessage = getHexMessage(publicKey);

          const signature = await VulticonnectWalletService.signCustomMessage(
            hexMessage,
            walletAddress
          );

          if (!signature || typeof signature !== "string") {
            throw new Error("Failed to sign message");
          }
          
          const token = await PolicyService.getAuthToken(
            hexMessage,
            signature,
            publicKey,
            chainCodeHex,
            derivePath
          );
          
          localStorage.setItem("authToken", token);
          setConnectedWallet(true);
          break;
        }

        default:
          throw new Error(`Chain ${chain} is currently not supported.`);
      }
    } catch (err) {
      console.error("Wallet connection error:", err);
      setError(err instanceof Error ? err.message : "Failed to connect wallet");
      setConnectedWallet(false);
    } finally {
      setIsConnecting(false);
    }
  };

  return (
    <div style={{ position: "relative" }}>
      <Button
        size="medium"
        styleType="primary"
        type="button"
        onClick={() => connectWallet(chain)}
        disabled={isConnecting}
      >
        {isConnecting ? "Connecting..." : connectedWallet ? "Connected" : "Connect Wallet"}
      </Button>
      {error && (
        <div style={{
          position: "absolute",
          top: "100%",
          left: 0,
          right: 0,
          color: "var(--color-error, #ff4d4f)",
          fontSize: "0.75rem",
          marginTop: "0.25rem",
          textAlign: "center"
        }}>
          {error}
        </div>
      )}
    </div>
  );
};

export default Wallet;
