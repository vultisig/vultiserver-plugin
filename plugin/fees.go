package plugin

import (
	"fmt"
	"math"
	"math/big"

	"github.com/vultisig/vultiserver-plugin/internal/types"
)

const AMOUNT_DENOMINATOR = 100_000
const PERCENTAGE_DENOMINATOR = 100

/**
 * Calculate fee amount based on plugin pricing policy
 */
func CalculateFeeAmount(pricingPolicy *types.PricingPolicy, taxBase *big.Int) (*big.Int, error) {
	amount := big.NewInt(0)
	// Assumptions:
	// token is ERC20 with 6 decimals
	// fee is taken on top of the base amount, not deducted from it
	if pricingPolicy.Metric == "FIXED" {
		amount.SetInt64(int64(math.Round(pricingPolicy.Amount * AMOUNT_DENOMINATOR)))
	} else if pricingPolicy.Metric == "PERCENTAGE" {
		amount = new(big.Int).Div(
			new(big.Int).Mul(
				taxBase,
				big.NewInt(int64(math.Round(pricingPolicy.Amount*AMOUNT_DENOMINATOR))),
			),
			new(big.Int).Mul(big.NewInt(AMOUNT_DENOMINATOR), big.NewInt(PERCENTAGE_DENOMINATOR)),
		)
	} else {
		return nil, fmt.Errorf("fee pricing policy metric not supported: %s", pricingPolicy.Metric)
	}

	return amount, nil
}
