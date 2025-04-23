package payroll

import "github.com/vultisig/vultiserver-plugin/plugin"

type PayrollPolicy struct {
	ChainID    []string           `json:"chain_id"`
	TokenID    []string           `json:"token_id"`
	Recipients []PayrollRecipient `json:"recipients"`
	Schedule   plugin.Schedule    `json:"schedule"`
}

type PayrollRecipient struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}
