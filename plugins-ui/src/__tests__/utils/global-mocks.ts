import { SchemaUtilsType, WidgetProps } from "@rjsf/utils";
import { afterEach, vi } from "vitest";

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

afterEach(() => {
  vi.clearAllMocks(); // Reset all mocked calls between tests
});
