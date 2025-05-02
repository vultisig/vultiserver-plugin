package common

import (
	"encoding/hex"
	"fmt"

	"github.com/vultisig/vultiserver-plugin/internal/types"
)

/**
 * Basic plugin pricing validator to be reused across services.
 * Verifier should also compare signed pricing matches general definition
 * in plugin->policy for the plugin it is signed for.
 */
func ValidatePluginPricingPolicy(pricing *types.PluginPricing, pluginType string) error {
	if pricing.PluginType != pluginType {
		return fmt.Errorf("pricing policy does not match plugin type, expected: %s, got: %s", pluginType, pricing.PluginType)
	}

	if pricing.ChainCodeHex == "" {
		return fmt.Errorf("pricing policy does not contain chain_code_hex")
	}

	if pricing.PublicKeyEcdsa == "" {
		return fmt.Errorf("pricing policy does not contain public_key_ecdsa")
	}

	if pricing.PublicKeyEddsa == "" {
		return fmt.Errorf("pricing policy does not contain public_key_eddsa")
	}

	pubKeyBytes, err := hex.DecodeString(pricing.PublicKeyEcdsa)
	if err != nil {
		return fmt.Errorf("invalid hex ecdsa encoding: %w", err)
	}
	isValidPublicKey := CheckIfPublicKeyIsValid(pubKeyBytes, true)
	if !isValidPublicKey {
		return fmt.Errorf("invalid public_key_ecdsa")
	}

	pubKeyBytes, err = hex.DecodeString(pricing.PublicKeyEddsa)
	if err != nil {
		return fmt.Errorf("invalid hex eddsa encoding: %w", err)
	}
	isValidPublicKey = CheckIfPublicKeyIsValid(pubKeyBytes, false)
	if !isValidPublicKey {
		return fmt.Errorf("invalid public_key_eddsa")
	}

	return nil
}
