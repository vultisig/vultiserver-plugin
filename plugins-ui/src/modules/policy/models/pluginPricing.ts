export type PluginPricing = {
  id: string;
  public_key_ecdsa: string;
  public_key_eddsa: string;
  plugin_type: string;
  is_ecdsa: boolean,
  chain_code_hex: string,
  derive_path: string,
  signature: string;
  policy: PluginPricingPolicy;
}

type PluginPricingPolicy = {
  type: string;
  frequency?: string;
  amount: number;
  metric: string;
}
