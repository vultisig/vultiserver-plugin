import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { RJSFSchema, TitleFieldProps } from "@rjsf/utils";
import { TitleFieldTemplate } from "@/modules/policy/components/policy-title/PolicyTitle";
import { USDC_TOKEN, WETH_TOKEN } from "@/modules/shared/data/tokens";

const schema: RJSFSchema = {
  title: "DCA Plugin Policy",
};

const mockRegistry: TitleFieldProps["registry"] = {
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
  } as any,
  translateString: () => "string",
  formContext: {},
};

describe("TitleFieldTemplate", () => {
  it("should render default title template if no tokens are provided in ui:options", () => {
    render(
      <TitleFieldTemplate
        id="title"
        title="Title"
        schema={schema}
        registry={mockRegistry}
      />
    );

    const header = screen.getByTestId("title");
    expect(header).toBeInTheDocument();
    expect(header).toHaveTextContent("Title");
  });

  it("should render custom title template if tokens are provided in ui:options", () => {
    const registry = {
      ...mockRegistry,
      ...{
        formContext: {
          sourceTokenId: WETH_TOKEN,
          destinationTokenId: USDC_TOKEN,
          editing: true,
        },
      },
    };
    render(
      <TitleFieldTemplate
        id="title"
        title="Title"
        schema={schema}
        registry={registry}
      />
    );

    const header = screen.getByTestId("title");
    expect(header).toBeInTheDocument();
    expect(header).not.toHaveTextContent("Title");
  });
});
