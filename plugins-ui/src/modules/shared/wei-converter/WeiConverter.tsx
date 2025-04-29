import React, { useEffect, useState } from "react";
import { WidgetProps } from "@rjsf/utils";
import { ethers } from "ethers";
import { supportedTokens } from "../data/tokens";
import "./WeiConverter.css";

const DEBOUNCE_DELAY = 500;

const WeiConverter = (props: WidgetProps) => {
  const { value, readonly, onChange, formContext, schema } = props;

  const selectedToken = formContext?.sourceTokenId;
  const [inputValue, setInputValue] = useState("");

  useEffect(() => {
    if (!value || !schema.pattern) {
      setInputValue("");
      return;
    }
    if (!readonly) return;
    try {
      const decimals = supportedTokens[selectedToken]?.decimals;
      const formattedValue = ethers.formatUnits(value, decimals);
      setInputValue(formattedValue);
    } catch (error) {
      console.error(error);
    }
  }, [value, selectedToken, readonly]);

  useEffect(() => {
    const timeout = setTimeout(() => {
      if (!inputValue || !schema.pattern) {
        onChange("");
        return;
      }

      try {
        const decimals = supportedTokens[selectedToken]?.decimals;
        const convertedValue = ethers
          .parseUnits(inputValue, decimals)
          .toString();
        onChange(convertedValue);
      } catch (error) {
        console.error(error);
      }
    }, DEBOUNCE_DELAY);

    return () => clearTimeout(timeout);
  }, [inputValue, selectedToken]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(e.target.value);
  };

  return (
    <>
      <input
        id="wei"
        type="number"
        value={inputValue}
        readOnly={readonly}
        onChange={handleChange}
        data-testid="wei-converter"
      />
    </>
  );
};

export default WeiConverter;
