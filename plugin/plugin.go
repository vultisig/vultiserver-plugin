package plugin

import (
	"context"
	"embed"

	"github.com/vultisig/mobile-tss-lib/tss"
	"github.com/vultisig/vultiserver-plugin/internal/types"
)

// PluginConfigField represents a single configuration field
type PluginConfigField struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`  // "string", "number", "boolean", "select"
	Label       string      `json:"label"` // Human-readable label
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Options     []string    `json:"options,omitempty"`     // For select type
	Placeholder string      `json:"placeholder,omitempty"` // Optional placeholder text
	Description string      `json:"description,omitempty"` // Help text
}

// PluginConfig defines the configuration schema for a plugin
type PluginConfig struct {
	Fields []PluginConfigField `json:"fields"`
}

type Plugin interface {
	// Configuration methods
	GetConfigSchema() PluginConfig
	ValidateConfig(config map[string]interface{}) error

	// Core plugin methods
	FrontendSchema() embed.FS
	ValidatePluginPolicy(policyDoc types.PluginPolicy) error
	ProposeTransactions(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error)
	ValidateProposedTransactions(policy types.PluginPolicy, txs []types.PluginKeysignRequest) error
	SigningComplete(ctx context.Context, signature tss.KeysignResponse, signRequest types.PluginKeysignRequest, policy types.PluginPolicy) error
}
