import { PluginProgress } from "@/modules/policy/models/policy";
import { RJSFSchema, SchemaUtilsType, WidgetProps } from "@rjsf/utils";
import { vi } from "vitest";

export const mockRegistry = {
  fields: {},
  widgets: {},
  rootSchema: {},
  templates: {
    ArrayFieldTemplate: () => null,
    ArrayFieldDescriptionTemplate: () => null,
    ArrayFieldItemTemplate: () => null,
    ArrayFieldTitleTemplate: () => null,
    BaseInputTemplate: () => null,
    ButtonTemplates: {
      AddButton: () => null,
      MoveDownButton: () => null,
      MoveUpButton: () => null,
      RemoveButton: () => null,
      SubmitButton: () => null,
      CopyButton: () => null,
    },
    DescriptionFieldTemplate: () => null,
    ErrorListTemplate: () => null,
    FieldErrorTemplate: () => null,
    FieldHelpTemplate: () => null,
    FieldTemplate: () => null,
    ObjectFieldTemplate: () => null,
    TitleFieldTemplate: () => null,
    WrapIfAdditionalTemplate: () => null,
    UnsupportedFieldTemplate: () => null,
  },
  schemaUtils: {
    getDisplayLabel: () => true,
  } as Partial<SchemaUtilsType>,
  translateString: () => "string",
  formContext: {},
};

export const mockedDCAPolicy = {
  form: {
    plugin_type: "dca",
    plugin_version: "0.0.1",
    policy_version: "0.0.1",
    schema: {
      properties: {
        chain_id: {
          default: "1",
          type: "string",
        },
        destination_token_id: {
          default: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
          title: "I want to buy",
          type: "string",
        },
        price_range: {
          items: {
            required: ["title"],
            type: "object",
          },
          properties: {
            max: {
              pattern: "^(?!0$)(?!0+\\.0*$)[0-9]+(\\.[0-9]+)?$",
              type: "string",
            },
            min: {
              pattern: "^(?!0$)(?!0+\\.0*$)[0-9]+(\\.[0-9]+)?$",
              title: "Price Range (optional)",
              type: "string",
            },
          },
          type: "object",
        },
        schedule: {
          dependencies: {
            frequency: {
              oneOf: [
                {
                  properties: {
                    frequency: {
                      enum: ["minutely"],
                    },
                    interval: {
                      pattern: "^(1[5-9]|[2-9][0-9]+)(\\.[0-9]+)?$",
                      type: "string",
                    },
                  },
                },
                {
                  properties: {
                    frequency: {
                      enum: ["hourly", "daily", "weekly", "monthly"],
                    },
                    interval: {
                      pattern: "^(?!0$)(?!0+\\.0*$)[0-9]+(\\.[0-9]+)?$",
                      type: "string",
                    },
                  },
                },
              ],
            },
          },
          items: {
            type: "object",
          },
          properties: {
            frequency: {
              default: "minutely",
              enum: ["minutely", "hourly", "daily", "weekly", "monthly"],
              title: "Time",
              type: "string",
            },
            interval: {
              title: "Every",
              type: "string",
            },
          },
          required: ["interval", "frequency"],
          type: "object",
        },
        source_token_id: {
          default: "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
          type: "string",
        },
        total_amount: {
          pattern: "^(?!0$)(?!0+\\.0*$)[0-9]+(\\.[0-9]+)?$",
          title: "I want to allocate",
          type: "string",
        },
        total_orders: {
          pattern: "^(?!0$)(?!0+\\.0*$)[0-9]+(\\.[0-9]+)?$",
          title: "Over (orders)",
          type: "string",
        },
      },
      required: [
        "total_amount",
        "source_token_id",
        "destination_token_id",
        "total_orders",
      ],
      title: "DCA Plugin Policy",
      type: "object",
    },
    uiSchema: {
      chain_id: {
        "ui:widget": "hidden",
      },
      destination_token_id: {
        "ui:options": {
          classNames: "input-background stacked-input",
        },
        "ui:widget": "TokenSelector",
      },
      price_range: {
        max: {
          "ui:options": {
            classNames: "input-background stacked-input",
            label: false,
            placeholder: "Max Price",
          },
          "ui:readonly": false,
          "ui:style": {
            display: "flex",
            flexDirection: "column",
            justifyContent: "flex-end",
          },
        },
        min: {
          "ui:options": {
            classNames: "input-background stacked-input",
            placeholder: "Min Price",
          },
          "ui:readonly": false,
          "ui:style": {
            display: "flex",
            flexDirection: "column",
            justifyContent: "flex-end",
          },
        },
        "ui:options": {
          classNames: "form-row",
          label: false,
        },
        "ui:order": ["min", "max"],
      },
      schedule: {
        frequency: {
          "ui:classNames": "input-background stacked-input",
          "ui:hideError": true,
          "ui:readonly": false,
          "ui:style": {
            display: "flex",
            flexDirection: "column",
          },
        },
        interval: {
          "ui:classNames": "input-background stacked-input",
          "ui:hideError": false,
          "ui:readonly": false,
          "ui:style": {
            display: "flex",
            flexDirection: "column",
          },
        },
        "ui:classNames": "form-row",
        "ui:hideError": true,
        "ui:options": {
          label: false,
        },
        "ui:order": ["interval", "frequency"],
      },
      source_token_id: {
        "ui:options": {
          classNames: "input-background stacked-input",
          label: false,
        },
        "ui:style": {
          boxSizing: "border-box",
          display: "inline-block",
          marginTop: "37px",
          verticalAlign: "top",
          width: "48%",
        },
        "ui:widget": "TokenSelector",
      },
      total_amount: {
        "ui:classNames": "input-background stacked-input",
        "ui:style": {
          boxSizing: "border-box",
          display: "inline-block",
          marginRight: "2%",
          verticalAlign: "top",
          width: "48%",
        },
        "ui:widget": "WeiConverter",
      },
      total_orders: {
        "ui:classNames": "input-background stacked-input",
      },
      "ui:description": "Set up configuration settings for DCA Plugin Policy",
      "ui:order": [
        "total_amount",
        "source_token_id",
        "destination_token_id",
        "schedule",
        "total_orders",
        "*",
      ],
      "ui:submitButtonOptions": {
        submitText: "Save policy",
      },
    },
  } as RJSFSchema,
  table: {
    columns: [
      {
        accessorKey: "pair",
        cellComponent: "TokenPair",
        header: "Pair",
      },
      {
        accessorKey: "sell",
        cellComponent: "TokenAmount",
        header: "Sell Total",
      },
      {
        accessorKey: "orders",
        header: "Total orders",
      },
      {
        accessorKey: "toBuy",
        cellComponent: "TokenName",
        header: "To buy",
      },
      {
        accessorKey: "orderInterval",
        header: "Order interval",
      },
      {
        accessorKey: "status",
        cellComponent: "ActiveStatus",
        header: "Active",
      },
    ],
    mapping: {
      orderInterval: "policy.schedule.interval, policy.schedule.frequency",
      orders: "policy.total_orders",
      pair: ["policy.source_token_id", "policy.destination_token_id"],
      policyId: "id",
      sell: ["policy.total_amount", "policy.source_token_id"],
      status: "active",
      toBuy: "policy.destination_token_id",
    },
  },
};

export const mockPluginPolicy = {
  id: "c5196498-7191-49c9-9236-e9d95e8470d9",
  public_key: "Public Key",
  is_ecdsa: false,
  public_key_ecdsa: "public_key_1_ecdsa",
  public_key_eddsa: "public_key_1_eddsa",
  chain_code_hex: "Chain code HEX",
  derive_path: "Derive path",
  plugin_version: "v0.1.0",
  policy_version: "v0.2.0",
  plugin_type: "Plugin type",
  signature: "Signature",
  policy: {
    chain_id: "1",
    schedule: {
      interval: "1",
      frequency: "weekly",
    },
    price_range: {
      max: "750",
      min: "500",
    },
    total_amount: "100000000",
    total_orders: "1",
    source_token_id: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
    destination_token_id: "0xB8c77482e45F1F44dE1745F52C74426C631bDD52",
  },
  active: true,
  plugin_id: "",
  progress: PluginProgress.InProgress,
};

export const mockPlugin = {
  id: "411ba072-df9b-4686-9ac8-d6e492394ba7",
  created_at: "2025-04-08T13:11:06.183848Z",
  updated_at: "2025-04-08T13:11:06.183848Z",
  type: "dca",
  title: "DCA Plugin",
  description: "Dollar cost averaging plugin automation",
  metadata: '{"foo": "bar"}',
  server_endpoint: "http://localhost:8081",
  pricing_id: "3d2e4b50-0213-4751-a72c-45935d957c3f",
};

export const getWidgetPropsMock = (value: unknown) => {
  return {
    id: "id",
    label: "TokenSelector",
    name: "TokenSelector",
    onBlur: vi.fn(),
    schema: {},
    options: {},
    onFocus: vi.fn(),
    registry: mockRegistry,
    value,
    onChange: vi.fn(),
  } as WidgetProps;
};

export const mockEventBus = {
  publish: vi.fn(),
  subscribe: vi.fn(),
  unsubscribe: vi.fn(),
};

vi.mock("@/utils/eventBus", () => ({
  publish: mockEventBus.publish,
  subscribe: mockEventBus.subscribe,
  unsubscribe: mockEventBus.unsubscribe,
}));
