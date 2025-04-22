import { RJSFSchema, UiSchema } from "@rjsf/utils";
import { ColumnDef } from "@tanstack/react-table";

export type Policy<
  T = string | number | boolean | string[] | null | undefined,
> = {
  [key: string]: T | Policy<T>;
};

export type PluginPolicy = {
  id: string;
  public_key: string;
  is_ecdsa: boolean;
  chain_code_hex: string;
  derive_path: string;
  plugin_id: string;
  plugin_version: string;
  policy_version: string;
  plugin_type: string;
  signature: string;
  policy: Policy;
  active: boolean;
  progress: string;
};

export type PluginPoliciesMap = {
  policies: PluginPolicy[];
  total_count: number;
};

export type TransactionHistory = {
  history: PolicyTransactionHistory[];
  total_count: number;
};

export type PolicyTableColumn = ColumnDef<unknown> & {
  accessorKey: string;
  header: string;
  cellComponent?: string;
  expandable?: boolean;
};

export type PolicyTransactionHistory = {
  id: string;
  updated_at: string;
  status: string;
};

export type PolicySchema = {
  form: {
    schema: RJSFSchema;
    uiSchema: UiSchema;
    plugin_version: string;
    policy_version: string;
    plugin_type: string;
  };
  table: {
    columns: PolicyTableColumn[];
    mapping: Record<string, string | string[]>;
  };
};
