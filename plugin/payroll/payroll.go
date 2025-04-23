package payroll

import (
	"embed"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultiserver-plugin/plugin"
	"github.com/vultisig/vultiserver-plugin/storage"
)

const (
	PluginType    = "payroll"
	pluginVersion = "0.0.1"
	policyVersion = "0.0.1"
)

//go:embed frontend
var frontend embed.FS

type Plugin struct {
	db           storage.DatabaseStorage
	nonceManager *plugin.NonceManager
	rpcClient    *ethclient.Client
	logger       logrus.FieldLogger
}

type PluginConfig struct {
	RpcURL string `mapstructure:"rpc_url" json:"rpc_url"`
}

func NewPlugin(db storage.DatabaseStorage, logger logrus.FieldLogger, rawConfig map[string]interface{}) (*Plugin, error) {
	var cfg PluginConfig
	if err := mapstructure.Decode(rawConfig, &cfg); err != nil {
		return nil, err
	}

	rpcClient, err := ethclient.Dial(cfg.RpcURL)
	if err != nil {
		return nil, err
	}

	return &Plugin{
		db:           db,
		rpcClient:    rpcClient,
		nonceManager: plugin.NewNonceManager(rpcClient),
		logger:       logger,
	}, nil
}

func (p *Plugin) FrontendSchema() ([]byte, error) {
	return frontend.ReadFile("./fronted/index.html") //TODO: jsonSchema not implemented yet
}

func (p *Plugin) GetNextNonce(address string) (uint64, error) {
	return p.nonceManager.GetNextNonce(address)
}
