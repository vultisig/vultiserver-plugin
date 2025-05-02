// networks currently supported in vulticonnect
export const networks = {
  Ethereum: "0x1",
  THORChain: "Thorchain_1",
  Solana: "Solana_mainnet-beta",
  Bitcoin: "0x1f96",
  BitcoinCash: "0x2710",
  Dash: "Dash_dash",
  DogeCoin: "0x7d0",
  DyDx: "dydx-1",
  GaiaChain: "cosmoshub-4",
  Kujira: "kaiyo-1",
  LiteCoin: "Litecoin_litecoin",
  MayaChain: "MayaChain-1",
  Osmosis: "osmosis-1",
}

export const networkList = Object.entries(networks).map(([name, id]) => ({ id, name }))
