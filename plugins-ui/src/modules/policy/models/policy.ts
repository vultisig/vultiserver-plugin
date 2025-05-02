import { RJSFSchema, UiSchema } from "@rjsf/utils";
import { ColumnDef } from "@tanstack/react-table";

export enum PluginProgress {
  InProgress = 'IN PROGRESS',
  Done = 'DONE',
};

export type Policy<
  T = string | number | boolean | string[] | null | undefined,
> = {
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
  progress: PluginProgress;
  signature: string;
  policy: Policy;
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
