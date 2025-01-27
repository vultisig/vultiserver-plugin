package plugin

import (
	"context"
	"embed"

	"github.com/labstack/echo/v4"
	"github.com/vultisig/mobile-tss-lib/tss"
	"github.com/vultisig/vultisigner/internal/types"
)

type Plugin interface {
	SignPluginMessages(c echo.Context) error
	ValidatePluginPolicy(policyDoc types.PluginPolicy) error
	ConfigurePlugin(c echo.Context) error
	Frontend() embed.FS

	ProposeTransactions(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error)
	ValidateTransactionProposal(policy types.PluginPolicy, txs []types.PluginKeysignRequest) error

	GetNextNonce(address string) (uint64, error)
	SigningComplete(ctx context.Context, signature tss.KeysignResponse, signRequest types.PluginKeysignRequest) error
}
