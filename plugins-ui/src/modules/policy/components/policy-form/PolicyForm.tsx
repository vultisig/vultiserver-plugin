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
import { v4 as uuidv4 } from "uuid";

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
    "minutes-validation": (value: string) => {
      try {
        const valueAsInt = parseInt(value);
        return valueAsInt >= 15 && valueAsInt <= 60;
      } catch {
        return false;
      }
    },
    "hours-validation": (value: string) => {
      try {
        const valueAsInt = parseInt(value);
        return valueAsInt >= 1 && valueAsInt <= 23;
      } catch {
        return false;
      }
    },
    "days-validation": (value: string) => {
      try {
        const valueAsInt = parseInt(value);
        return valueAsInt >= 1 && valueAsInt <= 31;
      } catch {
        return false;
      }
    },
    "weeks-validation": (value: string) => {
      try {
        const valueAsInt = parseInt(value);
        return valueAsInt >= 1 && valueAsInt <= 52;
      } catch {
        return false;
      }
    },
    "months-validation": (value: string) => {
      try {
        const valueAsInt = parseInt(value);
        return valueAsInt >= 1 && valueAsInt <= 12;
      } catch {
        return false;
      }
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
        } catch (error) {
          if (error instanceof Error) {
            console.error("Failed to update policy:", error.message);
          }
        }

        return;
      }

      try {
        policy.id = uuidv4();
        addPolicy(policy).then((addedSuccessfully) => {
          if (!addedSuccessfully) return;

          setFormData(initialFormData); // Reset formData to initial state
          setFormKey((prevKey) => prevKey + 1); // Change key to force remount
          if (onSubmitCallback) {
            onSubmitCallback(policy);
          }
        });
      } catch (error) {
        if (error instanceof Error) {
          console.error("Failed to create policy:", error.message);
        }
      }
    }
  };

  const extractValueFromValidationError = (
    validationFieldPath: string[],
    object: Record<string, unknown>
  ) => {
    const currentField = validationFieldPath.shift();
    if (!currentField) return;
    if (validationFieldPath.length > 0) {
      return extractValueFromValidationError(
        validationFieldPath,
        object[currentField] as Record<string, unknown>
      );
    } else {
      return object[currentField];
    }
  };

  const transformPatternError = (error: RJSFValidationError) => {
    const value: unknown = extractValueFromValidationError(
      error.property?.split(".").filter((v) => !!v) || [],
      formData
    );
    const forbiddenSymbols = /,|{|}|!|#|'|"|~/;
    const numberPatterns = [
      // Note: Positive number pattern
      "^(?!0$)(?!0+\\.0*$)[0-9]+(\\.[0-9]+)?$",
      "^(1[5-9]|[2-9][0-9]+)(\\.[0-9]+)?$",
    ];
    // Note: We check if the matched pattern is validating Numbers
    if (
      forbiddenSymbols.test(`${value}`) &&
      numberPatterns.includes(error.params.pattern)
    ) {
      return "should not use forbidden symbols ,{}!#'\"~";
    }
    // Note: We check if the current selected field for Time is "minutely"
    if (
      error.params?.pattern === numberPatterns[1] &&
      parseInt(`${value}`) <= 15
    ) {
      return "should be a number equal or above 15";
    }

    // Note: We check if the matched pattern is validating positive numbers
    if (error.params.pattern === numberPatterns[0]) {
      return "should be a positive number";
    }
  };

  const transformErrors = (errors: RJSFValidationError[]) => {
    return errors.map((error) => {
      if (error.name === "pattern") {
        error.message = transformPatternError(error);
      }
      if (error.name === "required") {
        error.message = "this field is required";
      }
      if (error.name === "format") {
        switch (error.params.format) {
          case "evm-address":
            error.message = "should be valid EVM address";
            break;
          case "minutes-validation":
            error.message = "should be a positive number between 15 and 60";
            break;
          case "hours-validation":
            error.message = "should be a positive number between 1 and 24";
            break;
          case "days-validation":
            error.message = "should be a positive number between 1 and 31";
            break;
          case "weeks-validation":
            error.message = "should be a positive number between 1 and 52";
            break;
          case "months-validation":
            error.message = "should be a positive number between 1 and 12";
            break;
          default:
            break;
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
