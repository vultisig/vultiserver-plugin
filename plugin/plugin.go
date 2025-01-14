package plugin

import (
	"embed"

	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/labstack/echo/v4"
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
	SigningComplete(signedTx *gtypes.Transaction) error
}
