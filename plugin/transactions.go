package plugin

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/vultisig/vultiserver-plugin/internal/types"
	"github.com/vultisig/vultiserver-plugin/pkg/uniswap"
)

type RawTxData struct {
	TxHash     []byte
	RlpTxBytes []byte
	Type       string
}

func GenerateFeeTransactions(
	uniswapClient uniswap.Client,
	chainID *big.Int,
	signerAddress *common.Address,
	feeTokenAddress *common.Address,
	feeRecipientAddress *common.Address,
	taxBase *big.Int,
	nonceOffset uint64,
	pricingPolicy *types.PricingPolicy,
) ([]RawTxData, error) {
	var rawTxsData []RawTxData

	amount := big.NewInt(0)
	var err error
	switch {
	case pricingPolicy.Type == "PER_TX":
		amount, err = CalculateFeeAmount(pricingPolicy, taxBase)
		if err != nil {
			return rawTxsData, err
		}
	case pricingPolicy.Type == "RECURRING":
	case pricingPolicy.Type == "SINGLE":
	case pricingPolicy.Type == "FREE":
		return rawTxsData, fmt.Errorf("fee pricing policy type not implemented: ", pricingPolicy.Type)
	default:
		return rawTxsData, fmt.Errorf("fee pricing policy type not supported: ", pricingPolicy.Type)
	}

	txHash, rawTx, err := uniswapClient.ERC20Transfer(
		chainID,
		feeTokenAddress,
		signerAddress,
		feeRecipientAddress,
		amount,
		nonceOffset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make FEE transaction: %w", err)
	}

	feeTx := RawTxData{txHash, rawTx, "FEE"}
	rawTxsData = append(rawTxsData, feeTx)

	return rawTxsData, nil
}
