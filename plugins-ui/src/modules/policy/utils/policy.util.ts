import { PluginPolicy, Policy } from "../models/policy";

export const generatePolicy = (
  plugin_version: string,
  policy_version: string,
  plugin_type: string,
  policyId: string,
  policy: Policy
): PluginPolicy => {
  return {
    id: policyId,
    public_key: "",
    is_ecdsa: true,
    chain_code_hex: "",
    derive_path: "",
    plugin_id: "TODO",
    plugin_version,
    policy_version,
    plugin_type,
    signature: "",
    policy: convertToStrings(policy),
    active: true,
    progress: "",
  };
};

function convertToStrings(
  obj: Record<string, unknown>
): Record<string, string> {
  Object.keys(obj).forEach((k) => {
    if (obj[k] === null || obj[k] === undefined) {
      delete obj[k];
      return;
    }
    if (typeof obj[k] === "object" && obj !== null) {
      return convertToStrings(obj[k] as Record<string, unknown>);
    }
    if (Array.isArray(obj[k])) {
      return obj[k].map((item) => {
        if (typeof item === "object" && item !== null) {
          return convertToStrings(item);
        }
        return `${item}`;
      });
    }

    obj[k] = `${obj[k]}`;
  });

  return obj as Record<string, string>;
}

const getValueByPath = (obj: Record<string, any>, path: string) =>
  path.split(".").reduce((acc, part) => acc?.[part], obj);

export const mapTableColumnData = (
  value: PluginPolicy,
  mapping: Record<string, any>
) => {
  const obj: Record<string, any> = {};

  for (const [key, paths] of Object.entries(mapping)) {
    if (Array.isArray(paths)) {
      // If it's an array, extract multiple values and store as an array
      obj[key] = paths.map((path) => getValueByPath(value, path));
    } else if (paths.includes(",")) {
      // If it's a concatenated value, extract each and join them
      obj[key] = paths
        .split(",")
        .map((path: any) => getValueByPath(value, path.trim()))
        .join(" ");
    } else if (typeof paths === "string") {
      // If it's a direct string path, extract the value
      obj[key] = getValueByPath(value, paths);
    } else {
      // If it's a static value, assign it directly
      obj[key] = paths;
    }
  }

  return obj;
};

export const sortObjectAlphabetically = (obj: any): any => {
  if (Array.isArray(obj)) {
    return obj.map(sortObjectAlphabetically);
  } else if (obj && typeof obj === "object" && obj.constructor === Object) {
    return Object.fromEntries(
      Object.entries(obj)
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([key, value]) => [key, sortObjectAlphabetically(value)])
    );
  }
  return obj;
};
