import { cloneElement } from "react";
import { supportedTokens } from "../data/tokens";

interface ITokenImageProps {
  data: string;
}

const TokenImage = ({ data }: ITokenImageProps) => {
  return (
    <>
      {supportedTokens[data] &&
        cloneElement(supportedTokens[data].image, {
          width: 24,
          height: 24,
        })}
      &nbsp;
    </>
  );
};

export default TokenImage;
