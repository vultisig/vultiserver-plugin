import { networkList } from "@/modules/core/constants/networks"

type ChainSelectorProps = {
  chain: string;
  setChain: (chain: string) => void;
};

const ChainSelector = ({ chain, setChain }: ChainSelectorProps) => {
  return (
    <select
      style={{
        padding: "10px",
        borderRadius: "10px",
        background: "#abc",
        position: "relative",
      }}
      value={chain}
      onChange={(e) => setChain(e.target.value)}
    >
      {networkList.map((network) => (
        <option key={network.id} value={network.id}>
          {network.name}
        </option>
      ))}
    </select>
  );
};

export default ChainSelector;
