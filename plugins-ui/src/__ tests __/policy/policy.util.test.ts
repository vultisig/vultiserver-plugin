import { PluginProgress, Policy } from "@/modules/policy/models/policy";
import { generatePolicy } from "@/modules/policy/utils/policy.util";
import { describe, expect, it } from "vitest";

describe("generatePolicy", () => {
  it("should render Pause/Pause button for each policy depending on its state", () => {
    const policyData: Policy = {
      someNumberInput: 5,
      someBooleanInput: false,
      someStringInput: "text",
      someNestedInput: {
        someNumberInput: 5,
        someBooleanInput: false,
        someStringInput: "text",
      },
      someNullInput: null,
      someUndefinedInput: undefined,
    };

    const result = generatePolicy(
      "0.0.1",
      "0.0.1",
      "pluginType",
      "",
      policyData
    );

    expect(result).toStrictEqual({
      id: "",
      public_key_ecdsa: "",
      public_key_eddsa: "",
      plugin_version: "0.0.1",
      policy_version: "0.0.1",
      plugin_type: "pluginType",
      is_ecdsa: true,
      chain_code_hex: "",
      derive_path: "",
      active: true,
      progress: PluginProgress.InProgress,
      signature: "",
      policy: {
        someNumberInput: "5",
        someBooleanInput: "false",
        someStringInput: "text",
        someNestedInput: {
          someNumberInput: "5",
          someBooleanInput: "false",
          someStringInput: "text",
        },
      },
    });
  });
});
