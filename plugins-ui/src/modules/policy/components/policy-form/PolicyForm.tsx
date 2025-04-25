import Form, { IChangeEvent } from "@rjsf/core";
import { customizeValidator } from "@rjsf/validator-ajv8";
import "./PolicyForm.css";
import { generatePolicy } from "../../utils/policy.util";
import { PluginPolicy, PolicySchema } from "../../models/policy";
import { useEffect, useState } from "react";
import { usePolicies } from "../../context/PolicyProvider";
import { TitleFieldTemplate } from "../policy-title/PolicyTitle";
import TokenSelector from "@/modules/shared/token-selector/TokenSelector";
import TokenSelectorArray from "@/modules/shared/token-selector-array/TokenSelectorArray.tsx";
import WeiConverter from "@/modules/shared/wei-converter/WeiConverter";
import { RJSFValidationError } from "@rjsf/utils";

type PolicyFormProps = {
  data?: PluginPolicy;
  onSubmitCallback?: (data: PluginPolicy) => void;
};

const PolicyForm = ({ data, onSubmitCallback }: PolicyFormProps) => {
  const policyId = data?.id || "";

  const initialFormData = data ? data.policy : {}; // Define the initial form state
  const [formData, setFormData] = useState(initialFormData);
  const { addPolicy, updatePolicy, policySchemaMap, pluginType } =
    usePolicies();
  const [schema, setSchema] = useState<PolicySchema | null>(null);

  useEffect(() => {
    const savedSchema = policySchemaMap.get(pluginType);
    if (savedSchema) {
      setSchema(savedSchema);
    }
  }, [policySchemaMap]);

  const [formKey, setFormKey] = useState(0); // Changing this forces remount

  const onChange = (e: IChangeEvent) => {
    setFormData(e.formData);
  };

  const customFormats = {
    "evm-address": (value: string) => {
      const regex = /^0x[a-fA-F0-9]{40}$/g;
      return regex.test(value);
    },
  };

  const customValidator = customizeValidator({
    customFormats,
  });

  const onSubmit = async (submitData: IChangeEvent) => {
    if (schema?.form) {
      const policy: PluginPolicy = generatePolicy(
        schema.form.plugin_version,
        schema.form.policy_version,
        schema.form.plugin_type,
        policyId,
        submitData.formData
      );

      // check if form has policyId, this means we are editing policy
      if (policyId) {
        try {
          updatePolicy(policy).then((updatedSuccessfully) => {
            if (updatedSuccessfully && onSubmitCallback) {
              onSubmitCallback(policy);
            }
          });
        } catch (error: any) {
          console.error("Failed to update policy:", error.message);
        }

        return;
      }

      try {
        addPolicy(policy).then((addedSuccessfully) => {
          if (!addedSuccessfully) return;

          setFormData(initialFormData); // Reset formData to initial state
          setFormKey((prevKey) => prevKey + 1); // Change key to force remount
          if (onSubmitCallback) {
            onSubmitCallback(policy);
          }
        });
      } catch (error: any) {
        console.error("Failed to create policy:", error.message);
      }
    }
  };

  const transformErrors = (errors: RJSFValidationError[]) => {
    return errors.map((error) => {
      if (error.name === "pattern") {
        if (error.params.pattern === "^(1[5-9]|[2-9][0-9]+)(\\.[0-9]+)?$") {
          error.message = "should be a positive number above 15";
        }
        if (error.params.pattern === "^(?!0$)(?!0+\\.0*$)[0-9]+(\\.[0-9]+)?$") {
          error.message = "should be a positive number";
        }
      }
      if (error.name === "required") {
        error.message = "required";
      }
      if (error.name === "format") {
        if (error.params.format === "evm-address") {
          error.message = "should be valid EVM address";
        }
      }
      return error;
    });
  };

  return (
    <div className="policy-form">
      {schema && (
        <Form
          key={formKey} // Forces full re-render on reset
          idPrefix={pluginType}
          schema={schema.form.schema}
          uiSchema={schema.form.uiSchema}
          validator={customValidator}
          formData={formData}
          onSubmit={onSubmit}
          onChange={onChange}
          showErrorList={false}
          templates={{ TitleFieldTemplate }}
          widgets={{ TokenSelector, WeiConverter, TokenSelectorArray }}
          transformErrors={transformErrors}
          liveValidate={!!policyId}
          readonly={!!policyId}
          formContext={{
            editing: !!policyId,
            sourceTokenId:
              (formData.source_token_id as string) ||
              (formData.token_id as string[])?.at(0), // sourceTokenId is needed in WeiConverter/TitleFieldTemplate and probably on most of the existing plugins to get the rigth decimal places based on token address
            destinationTokenId: formData.destination_token_id as string, // destinationTokenId is needed in TitleFieldTemplate
          }}
        />
      )}
    </div>
  );
};

export default PolicyForm;
