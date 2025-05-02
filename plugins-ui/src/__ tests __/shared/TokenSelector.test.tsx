import TokenSelector from "@/modules/shared/token-selector/TokenSelector";
import { TitleFieldProps, WidgetProps } from "@rjsf/utils";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

describe("TokenSelector Component", () => {
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
  const mockProps: WidgetProps = {
    id: "token-selector",
    name: "token",
    label: "Token",
    schema: { type: "string" },
    uiSchema: {},
    formContext: {},
    registry: mockRegistry,
    options: {},
    value: "ETH",
    required: false,
    readonly: false,
    disabled: false,
    autofocus: false,
    rawErrors: [],
    onChange: vi.fn(),
    onBlur: vi.fn(),
    onFocus: vi.fn(),
    placeholder: "",
    hideError: false,
  };

  it("should open modal when button is clicked", () => {
    render(<TokenSelector {...mockProps} />);

    const button = screen.getByRole("button", { name: "Open modal" });
    expect(button).toBeInTheDocument();
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();

    fireEvent.click(button);

    const modal = screen.getByRole("dialog");
    expect(modal).toBeInTheDocument();
  });

  it("should dismiss modal when close button is clicked", () => {
    render(<TokenSelector {...mockProps} />);

    const openButton = screen.getByRole("button", { name: "Open modal" });
    fireEvent.click(openButton);

    const modal = screen.getByRole("dialog");
    expect(modal).toBeInTheDocument();

    const closeButton = screen.getByRole("button", { name: "Close modal" });
    fireEvent.click(closeButton);

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("should set selected token & dismiss modal when token is selected", () => {
    render(<TokenSelector {...mockProps} />);

    const openButton = screen.getByRole("button", { name: "Open modal" });
    fireEvent.click(openButton);

    const usdcItem = screen.getByText("USDC");
    fireEvent.click(usdcItem);

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(openButton).toHaveTextContent("USDC");
  });

  it("should filter out tokens that match search & shows all items when input is cleared", () => {
    render(<TokenSelector {...mockProps} />);

    const openButton = screen.getByRole("button", { name: "Open modal" });
    fireEvent.click(openButton);

    const input = screen.getByPlaceholderText("Search by token");
    fireEvent.change(input, { target: { value: "us" } });

    expect(screen.getByText("USDC")).toBeInTheDocument();
    expect(screen.getByText("USDT")).toBeInTheDocument();
    expect(screen.queryByText("UNI")).not.toBeInTheDocument();
    expect(screen.queryByText("AAVE")).not.toBeInTheDocument();

    fireEvent.change(input, { target: { value: "" } });

    expect(screen.getByText("USDC")).toBeInTheDocument();
    expect(screen.getByText("USDT")).toBeInTheDocument();
    expect(screen.getByText("UNI")).toBeInTheDocument();
    expect(screen.getByText("AAVE")).toBeInTheDocument();
  });

  it("should show message when no matches to the filter are found", () => {
    render(<TokenSelector {...mockProps} />);

    const openButton = screen.getByRole("button", { name: "Open modal" });
    fireEvent.click(openButton);

    const input = screen.getByPlaceholderText("Search by token");
    fireEvent.change(input, { target: { value: "Lorem ipsum" } });

    expect(screen.getByText("No matching options")).toBeInTheDocument();
  });

  it("should show message when no matches to the filter are found", () => {
    render(<TokenSelector {...mockProps} />);

    const openButton = screen.getByRole("button", { name: "Open modal" });
    fireEvent.click(openButton);

    const input = screen.getByPlaceholderText("Search by token");
    fireEvent.change(input, { target: { value: "Lorem ipsum" } });

    expect(openButton).toHaveTextContent("Unknown token");
  });
});
