declare global {
  interface Window {
    vultisig: {
      ethereum?: {
        request: (args: { method: string; params?: any[] }) => Promise<any>;
      };
      getVaults: () => Promise<Array<{
        publicKeyEcdsa: string;
        hexChainCode: string;
        [key: string]: any;
      }>>;
    };
    ethereum: any;
    thorchain: any;
    bitcoin: any;
  }
}

export {}; // Ensure this file is treated as a module
