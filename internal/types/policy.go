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

/**
 * Returns public key to sign transactions with, based on the chain for which the policy is signed
 */
func (p *PluginPolicy) GetPublicKey() string {
	if p.IsEcdsa {
		return p.PublicKeyEcdsa
	}
	return p.PublicKeyEddsa
}
