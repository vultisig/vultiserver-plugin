package uniswap

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Config struct {
	rpcClient          *ethclient.Client
	routerAddress      *common.Address
	swapGasLimit       uint64
	gasLimitBuffer     uint64 // TODO: remove
	slippagePercentage float64
	deadlineDuration   time.Duration
}

func NewConfig(rpcClient *ethclient.Client, routerAddress *common.Address, swapGasLimit, gasLimitBuffer uint64, slippagePercentage float64, deadlineDuration time.Duration) *Config {
	return &Config{
		rpcClient:          rpcClient,
		routerAddress:      routerAddress,
		swapGasLimit:       swapGasLimit,
		gasLimitBuffer:     gasLimitBuffer,
		slippagePercentage: slippagePercentage,
		deadlineDuration:   deadlineDuration,
	}
}
