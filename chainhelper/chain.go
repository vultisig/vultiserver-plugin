package chainhelper

import (
	"fmt"

	v1 "github.com/vultisig/commondata/go/vultisig/keysign/v1"
	"github.com/vultisig/vultisigner/walletcore/core"
)

type Chain string

const (
	THORChain   Chain = "THORChain"
	Solana      Chain = "Solana"
	Ethereum    Chain = "Ethereum"
	Avalanche   Chain = "Avalanche"
	BSC         Chain = "BSC"
	Bitcoin     Chain = "Bitcoin"
	Litecoin    Chain = "Litecoin"
	BitcoinCash Chain = "BitcoinCash"
	Dogecoin    Chain = "Dogecoin"
	Zcash       Chain = "Zcash"
	Dash        Chain = "Dash"
	MayaChain   Chain = "MayaChain"
	Arbitrum    Chain = "Arbitrum"
	Basechain   Chain = "Base"
	Optimism    Chain = "Optimism"
	Polygon     Chain = "Polygon"
	Blast       Chain = "Blast"
	CronosChain Chain = "CronosChain"
	Sui         Chain = "Sui"
	Polkadot    Chain = "Polkadot"
	Zksync      Chain = "Zksync"
	Kujira      Chain = "Kujira"
	Dydx        Chain = "Dydx"
	Cosmos      Chain = "Cosmos"
)

type ChainHelper interface {
	GetPreSignedImageHash(payload *v1.KeysignPayload) ([]string, error)
}

func NewChainHelper(chainName string) (ChainHelper, error) {
	switch Chain(chainName) {
	case Ethereum, BSC, Avalanche, Arbitrum, Basechain, Optimism, Polygon, Blast, CronosChain:
		return NewEVMChainHelper(getCoinType(Chain(chainName))), nil
	case Bitcoin, BitcoinCash, Litecoin, Dogecoin, Zcash, Dash:
		return NewUTXOChainHelper(getCoinType(Chain(chainName))), nil
	case THORChain:
		return NewTHORChainHelper(), nil
	// Add other chains as needed
	default:
		return nil, fmt.Errorf("unsupported chain: %s", chainName)
	}
}

func getCoinType(chain Chain) core.CoinType {
	switch chain {
	case Bitcoin:
		return core.CoinTypeBitcoin
	case BitcoinCash:
		return core.CoinTypeBitcoinCash
	case Litecoin:
		return core.CoinTypeLitecoin
	case Dogecoin:
		return core.CoinTypeDogecoin
	case Zcash:
		return core.CoinTypeZcash
	case Dash:
		return core.CoinTypeDash
	case THORChain:
		return core.CoinTypeTHORChain
	case Solana:
		return core.CoinTypeSolana
	case Ethereum:
		return core.CoinTypeEthereum
	case Avalanche:
		return core.CoinTypeAvalanche
	case BSC:
		return core.CoinTypeSmartChain
	case MayaChain:
		return core.CoinTypeTHORChain // Using THORChain since MayaChain not defined in core
	case Arbitrum:
		return core.CoinTypeArbitrum
	case Basechain:
		return core.CoinTypeBase
	case Optimism:
		return core.CoinTypeOptimism
	case Polygon:
		return core.CoinTypePolygon
	case Blast:
		return core.CoinTypeBlast
	case CronosChain:
		return core.CoinTypeCronos
	case Sui:
		return core.CoinTypeSui
	case Polkadot:
		return core.CoinTypePolkadot
	case Zksync:
		return core.CoinTypeZKSync
	case Kujira:
		return core.CoinTypeKujira
	case Dydx:
		return core.CoinTypeDydx
	case Cosmos:
		return core.CoinTypeCosmos
	default:
		return 0 // Using 0 instead of CoinTypeUnknown since it's not defined in core
	}
}
