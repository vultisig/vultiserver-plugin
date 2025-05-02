package payroll

import "github.com/vultisig/vultiserver-plugin/plugin"

type Policy struct {
	ChainID    []string        `json:"chain_id"`
	TokenID    []string        `json:"token_id"`
	Recipients []Recipient     `json:"recipients"`
	Schedule   plugin.Schedule `json:"schedule"`
}

type Recipient struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}
