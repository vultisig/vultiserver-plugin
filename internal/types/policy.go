package types

import (
	"encoding/json"
)

type PluginTriggerEvent struct {
	PolicyID string `json:"policy_id"`
}

type PluginPolicy struct {
	ID             string          `json:"id" validate:"required"`
	PublicKeyEcdsa string          `json:"public_key_ecdsa" validate:"required"`
	PublicKeyEddsa string          `json:"public_key_eddsa" validate:"required"`
	PluginVersion  string          `json:"plugin_version" validate:"required"`
	PolicyVersion  string          `json:"policy_version" validate:"required"`
	PluginType     string          `json:"plugin_type" validate:"required"`
	IsEcdsa        bool            `json:"is_ecdsa"`
	ChainCodeHex   string          `json:"chain_code_hex" validate:"required"`
	DerivePath     string          `json:"derive_path" validate:"required"`
	Active         bool            `json:"active"`
	Progress       string          `json:"progress" validate:"required"`
	Signature      string          `json:"signature" validate:"required"`
	Policy         json.RawMessage `json:"policy" validate:"required"`
}

type PluginPolicyPaginatedList struct {
	Policies   []PluginPolicy `json:"policies" validate:"required"`
	TotalCount int            `json:"total_count" validate:"required"`
}

type PayrollPolicy struct {
	ChainID    []string           `json:"chain_id"`
	TokenID    []string           `json:"token_id"`
	Recipients []PayrollRecipient `json:"recipients"`
	Schedule   Schedule           `json:"schedule"`
}

type DCAPolicy struct {
	ChainID            string     `json:"chain_id"`
	SourceTokenID      string     `json:"source_token_id"`
	DestinationTokenID string     `json:"destination_token_id"`
	TotalAmount        string     `json:"total_amount"`
	TotalOrders        string     `json:"total_orders"`
	Schedule           Schedule   `json:"schedule"`
	PriceRange         PriceRange `json:"price_range"`
}

type PayrollRecipient struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

type Schedule struct {
	Frequency string `json:"frequency"`
	Interval  string `json:"interval"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time,omitempty"`
}

type PriceRange struct {
	Min string `json:"min"`
	Max string `json:"max"`
}

/**
 * Returns public key to sign transactions with, based on the chain for which the policy is signed
 */
func (p *PluginPolicy) GetPublicKey() string {
	if p.IsEcdsa {
		return p.PublicKeyEcdsa
	}
	return p.PublicKeyEddsa
}
