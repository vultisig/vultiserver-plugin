import { supportedTokens } from "@/modules/shared/data/tokens";
import "./TokenPair.css";
import TokenImage from "../token-image/TokenImage";

type TokenPairProps = {
  data: [string, string];
};

const TokenPair = ({ data }: TokenPairProps) => {
  const [source_token_id, destination_token_id] = data;

  return (
    <div className="pair">
      <div className="token-icon">
        <TokenImage data={source_token_id} />
      </div>
      <div className="token-icon-right">
        <TokenImage data={source_token_id} />
      </div>
      {supportedTokens[source_token_id]?.name ||
        `Unknown token address: ${source_token_id}`}
      /
      {supportedTokens[destination_token_id]?.name ||
        `Unknown token address: ${destination_token_id}`}
    </div>
  );
};

export default TokenPair;
