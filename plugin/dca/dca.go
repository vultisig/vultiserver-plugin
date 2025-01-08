package dca

import (
	"embed"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	gcommon "github.com/ethereum/go-ethereum/common"
	"github.com/labstack/echo/v4"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
)

const (
	PLUGIN_TYPE = "dca"
)

type DCAPlugin struct {
	db storage.DatabaseStorage
}

func NewDCAPlugin(db storage.DatabaseStorage) *DCAPlugin {
	return &DCAPlugin{db}
}

func (p *DCAPlugin) SignPluginMessages(e echo.Context) error { return nil }

func (p *DCAPlugin) ValidatePluginPolicy(policyDoc types.PluginPolicy) error {
	if policyDoc.PluginType != PLUGIN_TYPE {
		return fmt.Errorf("policy does not match plugin type, expected: %s, got: %s", PLUGIN_TYPE, policyDoc.PluginType)
	}

	var dcaPolicy types.DCAPolicy
	if err := json.Unmarshal(policyDoc.Policy, &dcaPolicy); err != nil {
		return fmt.Errorf("failed to unmarshal DCA policy: %w", err)
	}

	mixedCaseTokenIn, err := gcommon.NewMixedcaseAddressFromString(dcaPolicy.TokenInID)
	if err != nil {
		return fmt.Errorf("invalid tokenIn address: %s", dcaPolicy.TokenInID)
	}
	if strings.ToLower(dcaPolicy.TokenInID) != dcaPolicy.TokenInID {
		if !mixedCaseTokenIn.ValidChecksum() {
			return fmt.Errorf("invalid tokenIn address checksum: %s", dcaPolicy.TokenInID)
		}
	}

	mixedCaseTokenOut, err := gcommon.NewMixedcaseAddressFromString(dcaPolicy.TokenOutID)
	if err != nil {
		return fmt.Errorf("invalid tokenOut address: %s", dcaPolicy.TokenOutID)
	}
	if strings.ToLower(dcaPolicy.TokenOutID) != dcaPolicy.TokenOutID {
		if !mixedCaseTokenOut.ValidChecksum() {
			return fmt.Errorf("invalid tokenOut address checksum: %s", dcaPolicy.TokenOutID)
		}
	}

	if dcaPolicy.TokenInID == dcaPolicy.TokenOutID {
		return fmt.Errorf("tokenIn and tokenOut addresses are the same")
	}

	if dcaPolicy.Amount == "" {
		return fmt.Errorf("amount is required")
	}
	amount, ok := new(big.Int).SetString(dcaPolicy.Amount, 10)
	if !ok {
		return fmt.Errorf("invalid amount %s", dcaPolicy.Amount)
	}
	if amount.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	if dcaPolicy.SlippagePercentage == "" {
		return fmt.Errorf("slippage percentage is required")
	}
	slippage, err := strconv.ParseFloat(dcaPolicy.SlippagePercentage, 64)
	if err != nil {
		return fmt.Errorf("invalid slippage percentage %s", dcaPolicy.SlippagePercentage)
	}
	if slippage <= 0 || slippage > 100 {
		return fmt.Errorf("slippage percentage must be between 0 and 100 %s", dcaPolicy.SlippagePercentage)
	}

	if dcaPolicy.ChainID == "" {
		return fmt.Errorf("chain id is required")
	}

	return nil
}

func (p *DCAPlugin) ConfigurePlugin(e echo.Context) error { return nil }

func (p *DCAPlugin) ProposeTransactions(policy types.PluginPolicy) ([]types.PluginKeysignRequest, error) {
	return nil, nil
}

func (p *DCAPlugin) ValidateTransactionProposal(policy types.PluginPolicy, txs []types.PluginKeysignRequest) error {
	return nil
}

func (p *DCAPlugin) Frontend() embed.FS { return embed.FS{} }
