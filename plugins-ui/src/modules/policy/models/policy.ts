import { RJSFSchema } from "@rjsf/utils";

export type Policy<T = string | number | boolean | null | undefined> = {
  [key: string]: T | Policy<T>;
};

export type PluginPolicy = {
  id: string;
  public_key_ecdsa: string;
  public_key_eddsa: string;
  plugin_version: string;
  policy_version: string;
  plugin_type: string;
  is_ecdsa: boolean;
  chain_code_hex: string;
  derive_path: string;
  active: boolean;
  signature: string;
  policy: Policy;
};

export type PolicyTransactionHistory = {
  id: string;
  updated_at: string;
  status: string;
};

export type PolicySchema = {
  form: {
    schema: RJSFSchema;
    uiSchema: {};
    plugin_version: string;
    policy_version: string;
    plugin_type: string;
  };
  table: {
    columns: [];
    mapping: {};
  };
};
