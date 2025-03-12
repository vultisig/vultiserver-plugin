import { PluginPolicy, Policy } from "../models/policy";
// const PUBLIC_KEY = import.meta.env.VITE_PUBLIC_KEY;

export const generatePolicy = (
  plugin_type: string,
  policyId: string,
  policy: Policy
): PluginPolicy => {
  return {
    id: policyId,
    public_key: "", // PUBLIC_KEY
    plugin_id: "dca",
    plugin_version: "0.0.1",
    policy_version: "0.0.1",
    plugin_type,
    signature: "",
    policy: convertToStrings(policy),
    active: true,
  };
};

function convertToStrings<T extends Record<string, any>>(
  obj: T
): Record<string, string> {
  return Object.fromEntries(
    Object.entries(obj).map(([key, value]) => [
      key,
      typeof value === "object" && value !== null
        ? convertToStrings(value)
        : String(value),
    ])
  ) as Record<string, string>;
}
