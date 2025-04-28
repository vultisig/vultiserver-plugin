package types

import "encoding/json"

type PluginPricingCreateDto struct {
	PublicKeyEcdsa string          `json:"public_key_ecdsa" validate:"required"`
	PublicKeyEddsa string          `json:"public_key_eddsa" validate:"required"`
	PluginType     string          `json:"plugin_type" validate:"required"`
	IsEcdsa        bool            `json:"is_ecdsa"`
	ChainCodeHex   string          `json:"chain_code_hex" validate:"required"`
	DerivePath     string          `json:"derive_path" validate:"required"`
	Signature      string          `json:"signature" validate:"required"`
	Policy         json.RawMessage `json:"policy" validate:"required"`
}

type PluginPricing struct {
	ID             string          `json:"id" validate:"required"`
	PublicKeyEcdsa string          `json:"public_key_ecdsa" validate:"required"`
	PublicKeyEddsa string          `json:"public_key_eddsa" validate:"required"`
	PluginType     string          `json:"plugin_type" validate:"required"`
	IsEcdsa        bool            `json:"is_ecdsa"`
	ChainCodeHex   string          `json:"chain_code_hex" validate:"required"`
	DerivePath     string          `json:"derive_path" validate:"required"`
	Signature      string          `json:"signature" validate:"required"`
	Policy         json.RawMessage `json:"policy" validate:"required"`
}

type PricingPolicy struct {
	Type string `json:"type"`
	// Frequency string  `json:"frequency"`
	Amount float64 `json:"amount"`
	Metric string  `json:"metric"`
}
