package dca

import "github.com/vultisig/vultisigner/plugin"

type DCAPolicy struct {
	ChainID            string          `json:"chain_id"`
	SourceTokenID      string          `json:"source_token_id"`
	DestinationTokenID string          `json:"destination_token_id"`
	TotalAmount        string          `json:"total_amount"`
	TotalOrders        string          `json:"total_orders"`
	Schedule           plugin.Schedule `json:"schedule"`
	PriceRange         PriceRange      `json:"price_range"`
}

type PriceRange struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

type RawTxData struct {
	TxHash     []byte
	RlpTxBytes []byte
	Type       string
}

type DCAPluginConfig struct {
	RpcURL  string `mapstructure:"rpc_url" json:"rpc_url"`
	Uniswap struct {
		V2Router string `mapstructure:"v2_router" json:"v2_router"`
		Deadline int64  `mapstructure:"deadline" json:"deadline"`
	} `mapstructure:"uniswap" json:"uniswap"`
}
