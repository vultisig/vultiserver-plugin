import Form from "@rjsf/core";
import { RJSFSchema } from "@rjsf/utils";
import validator from "@rjsf/validator-ajv8";
import { useEffect, useState } from "react";
import "./PluginConfigForm.css";

interface PluginConfigFormProps {
  configSchema: {
    fields: Array<{
      name: string;
      type: string;
      label: string;
      required: boolean;
      default?: any;
      options?: string[];
      placeholder?: string;
      description?: string;
    }>;
  };
  onSubmit: (config: any) => void;
}

const PluginConfigForm: React.FC<PluginConfigFormProps> = ({ configSchema, onSubmit }) => {
  const [formSchema, setFormSchema] = useState<RJSFSchema | null>(null);
  const [uiSchema, setUiSchema] = useState<any>({});

  useEffect(() => {
    // Convert plugin config schema to RJSF schema
    const properties: any = {};
    const required: string[] = [];
    const newUiSchema: any = {};

    configSchema.fields.forEach((field) => {
      properties[field.name] = {
        type: convertType(field.type),
        title: field.label,
        description: field.description,
      };

      if (field.type === "select" && field.options) {
        properties[field.name].enum = field.options;
      }

      if (field.default !== undefined) {
        properties[field.name].default = field.default;
      }

      if (field.required) {
        required.push(field.name);
      }

      // Add UI customizations
      newUiSchema[field.name] = {
        "ui:placeholder": field.placeholder,
        "ui:help": field.description,
      };
    });

    setFormSchema({
      type: "object",
      required,
      properties,
    });

    setUiSchema(newUiSchema);
  }, [configSchema]);

  const handleSubmit = ({ formData }: any) => {
    onSubmit(formData);
  };

  if (!formSchema) {
    return null;
  }

  return (
    <div className="plugin-config-form">
      <Form
        schema={formSchema}
        uiSchema={uiSchema}
        validator={validator}
        onSubmit={handleSubmit}
        showErrorList={false}
      />
    </div>
  );
};

// Helper function to convert plugin config types to JSON Schema types
const convertType = (type: string): string => {
  switch (type) {
    case "number":
      return "number";
    case "boolean":
      return "boolean";
    case "select":
    case "string":
    default:
      return "string";
  }
};

export default PluginConfigForm; 