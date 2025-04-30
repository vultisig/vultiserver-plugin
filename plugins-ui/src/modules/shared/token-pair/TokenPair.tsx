import { supportedTokens } from "@/modules/shared/data/tokens";
import { cloneElement } from "react";

import "./TokenPair.css";

export type TokenPairProps = {
  data: [string, string];
};

const TokenPair = ({ data }: TokenPairProps) => {
  const [token, secondToken] = data;

  return (
    <div className="pair">
      <div className="token-icon">
        {supportedTokens[token] &&
          cloneElement(supportedTokens[token].image, {
            width: 24,
            height: 24,
          })}
        <span>
          {supportedTokens[token]?.name || `Unknown token address: ${token}`}
        </span>
      </div>

      {secondToken && (
        <div className="token-icon token-icon-bottom">
          {supportedTokens[secondToken] &&
            cloneElement(supportedTokens[secondToken].image, {
              width: 24,
              height: 24,
            })}
          <span>
            {supportedTokens[secondToken]?.name ||
              `Unknown token address: ${secondToken}`}
          </span>
        </div>
      )}
    </div>
  );
};

export default TokenPair;
