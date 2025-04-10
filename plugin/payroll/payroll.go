package payroll

import (
	"embed"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/plugin"
	"github.com/vultisig/vultisigner/storage"
)

const (
	PluginType    = "payroll"
	pluginVersion = "0.0.1"
	policyVersion = "0.0.1"
)

//go:embed frontend
var frontend embed.FS

type PayrollPlugin struct {
	db           storage.DatabaseStorage
	nonceManager *plugin.NonceManager
	rpcClient    *ethclient.Client
	logger       logrus.FieldLogger
}

type PayrollPluginConfig struct {
	rpcURL string `mapstructure:"rpc_url" json:"rpc_url"`
}

func NewPayrollPlugin(db storage.DatabaseStorage, logger logrus.FieldLogger, rawConfig map[string]interface{}) (*PayrollPlugin, error) {
	var cfg PayrollPluginConfig
	if err := mapstructure.Decode(rawConfig, &cfg); err != nil {
		return nil, err
	}

	rpcClient, err := ethclient.Dial(cfg.rpcURL)
	if err != nil {
		return nil, err
	}

	return &PayrollPlugin{
		db:           db,
		rpcClient:    rpcClient,
		nonceManager: plugin.NewNonceManager(rpcClient),
		logger:       logger,
	}, nil
}

func (p *PayrollPlugin) FrontendSchema() embed.FS {
	return frontend
}

func (p *PayrollPlugin) GetNextNonce(address string) (uint64, error) {
	return p.nonceManager.GetNextNonce(address)
}
