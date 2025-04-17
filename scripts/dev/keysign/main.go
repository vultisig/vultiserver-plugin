package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/types"
)

var vaultName string
var stateDir string

func main() {
	flag.StringVar(&vaultName, "vault", "", "vault name")
	flag.StringVar(&stateDir, "state-dir", "", "state directory")
	flag.Parse()

	if vaultName == "" {
		panic("vault name is required")
	}

	if stateDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		stateDir = filepath.Join(homeDir, ".vultiserver", "vaults")
	}

	vaultFile, err := os.Open(filepath.Join(stateDir, vaultName, "public_key"))
	if err != nil {
		panic(err)
	}

	pubKey, err := os.ReadFile(vaultFile.Name())
	if err != nil {
		panic(err)
	}

	serverConfig, err := config.ReadConfig("config-verifier")
	if err != nil {
		panic(err)
	}

	pluginConfig, err := config.ReadConfig("config-plugin")
	if err != nil {
		panic(err)
	}

	// VERIFIER

	verifierHost := fmt.Sprintf("http://%s:%d", serverConfig.Verifier.Host, serverConfig.Verifier.Port)
	fmt.Printf("Verifier keysign - %s/vault/sign", verifierHost)

	signMsgRequest := &types.PluginKeysignRequest{
		KeysignRequest: types.KeysignRequest{
			SessionID:        uuid.New().String(),
			PublicKey:        string(pubKey),
			IsECDSA:          true,
			VaultPassword:    "your-secure-password",
			HexEncryptionKey: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			DerivePath:       "m/44'/60'/0'/0/0",
			StartSession:     false,
			Messages:         []string{""},
		},
		Transaction: "ec8085056acafea38301286a94c02aaa39b223fe8d0a0e5c4f27ead9083c756cc28398968084d0e30db0018080",
		PolicyID:    uuid.New().String(),
	}

	reqBytes, err := json.Marshal(signMsgRequest)
	if err != nil {
		panic(err)
	}
	resp, err := http.Post(fmt.Sprintf("%s/vault/sign", verifierHost), "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		panic(err)
	}
	fmt.Printf(" - %d\n", resp.StatusCode)

	time.Sleep(3 * time.Second)

	// PLUGIN

	pluginHost := fmt.Sprintf("http://%s:%d", pluginConfig.Plugin.Host, pluginConfig.Plugin.Port)
	fmt.Printf("Plugin keysign - %s/vault/sign", pluginHost)

	signMsgRequest.StartSession = true
	signMsgRequest.Parties = []string{common.PluginPartyID, common.VerifierPartyID}

	reqBytes, err = json.Marshal(signMsgRequest)
	if err != nil {
		panic(err)
	}
	resp, err = http.Post(fmt.Sprintf("%s/vault/sign", pluginHost), "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		panic(err)
	}
	fmt.Printf(" - %d\n", resp.StatusCode)
}
