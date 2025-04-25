package types

import "encoding/json"

type PluginTriggerEvent struct {
	PolicyID string `json:"policy_id"`
}

type PluginPolicy struct {
	ID            string          `json:"id" validate:"required"`
	PublicKey     string          `json:"public_key" validate:"required"`
	IsEcdsa       bool            `json:"is_ecdsa" validate:"required"`
	ChainCodeHex  string          `json:"chain_code_hex" validate:"required"`
	DerivePath    string          `json:"derive_path" validate:"required"`
	PluginID      string          `json:"plugin_id" validate:"required"`
	PluginVersion string          `json:"plugin_version" validate:"required"`
	PolicyVersion string          `json:"policy_version" validate:"required"`
	PluginType    string          `json:"plugin_type" validate:"required"`
	Signature     string          `json:"signature" validate:"required"`
	Policy        json.RawMessage `json:"policy" validate:"required"`
	Active        bool            `json:"active" validate:"required"`
	Progress      string          `json:"progress" validate:"required"`
}
