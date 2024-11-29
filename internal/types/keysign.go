package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	v1 "github.com/vultisig/commondata/go/vultisig/keysign/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

type KeysignRequest struct {
	PublicKey        string             `json:"public_key"`         // public key, used to identify the backup file
	Payload          *v1.KeysignPayload `json:"payload"`            // payload
	SessionID        string             `json:"session"`            // Session ID , it should be an UUID
	HexEncryptionKey string             `json:"hex_encryption_key"` // Hex encryption key, used to encrypt the keysign messages
	DerivePath       string             `json:"derive_path"`        // Derive Path
	IsECDSA          bool               `json:"is_ecdsa"`           // indicate use ECDSA or EDDSA key to sign the messages
	VaultPassword    string             `json:"vault_password"`     // password used to decrypt the vault file
	StartSession     bool               `json:"start_session"`      // indicate start a new session or not
	Parties          []string           `json:"parties"`            // parties to join the session
}

// IsValid checks if the keysign request is valid
func (r KeysignRequest) IsValid() error {
	if r.PublicKey == "" {
		return errors.New("invalid public key ECDSA")
	}
	/*if len(r.Messages) == 0 {
		return errors.New("invalid messages")
	}*/
	if r.SessionID == "" {
		return errors.New("invalid session")
	}
	if r.HexEncryptionKey == "" {
		return errors.New("invalid hex encryption key")
	}
	if r.DerivePath == "" {
		return errors.New("invalid derive path")
	}

	return nil
}

type PluginKeysignRequest struct {
	KeysignRequest KeysignRequest `json:"keysign_request"`
	//Transactions []string `json:"transaction_hash"`
	PluginID string `json:"plugin_id"`
	PolicyID string `json:"policy_id"`
}

// Add a custom MarshalJSON method to KeysignRequest
func (k KeysignRequest) MarshalJSON() ([]byte, error) {

	// Use protojson for the payload
	marshaler := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}

	// Marshal the payload first
	payloadBytes, err := marshaler.Marshal(k.Payload)
	if err != nil {
		return nil, err
	}

	// Parse it back to a raw message
	var rawPayload json.RawMessage
	if err := json.Unmarshal(payloadBytes, &rawPayload); err != nil {
		return nil, err
	}

	// Create the final structure
	finalStruct := struct {
		PublicKey        string          `json:"public_key"`
		Payload          json.RawMessage `json:"payload"`
		SessionID        string          `json:"session"`
		HexEncryptionKey string          `json:"hex_encryption_key"`
		DerivePath       string          `json:"derive_path"`
		IsECDSA          bool            `json:"is_ecdsa"`
		VaultPassword    string          `json:"vault_password"`
		StartSession     bool            `json:"start_session"`
		Parties          []string        `json:"parties"`
	}{
		PublicKey:        k.PublicKey,
		Payload:          rawPayload,
		SessionID:        k.SessionID,
		HexEncryptionKey: k.HexEncryptionKey,
		DerivePath:       k.DerivePath,
		IsECDSA:          k.IsECDSA,
		VaultPassword:    k.VaultPassword,
		StartSession:     k.StartSession,
		Parties:          k.Parties,
	}

	return json.Marshal(finalStruct)
}

func parseNonce(nonceStr string) (int64, error) {
	nonce, err := strconv.ParseInt(nonceStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return nonce, nil
}

// Custom unmarshaler for KeysignRequest
func (k *KeysignRequest) UnmarshalJSON(data []byte) error {
	// Temporary struct for initial unmarshal
	var temp struct {
		PublicKey        string          `json:"public_key"`
		Payload          json.RawMessage `json:"payload"`
		SessionID        string          `json:"session"`
		HexEncryptionKey string          `json:"hex_encryption_key"`
		DerivePath       string          `json:"derive_path"`
		IsECDSA          bool            `json:"is_ecdsa"`
		VaultPassword    string          `json:"vault_password"`
		StartSession     bool            `json:"start_session"`
		Parties          []string        `json:"parties"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Temporary struct for ethereum specific with string nonce
	type tempEthereumSpecific struct {
		MaxFeePerGasWei string `json:"max_fee_per_gas_wei"`
		PriorityFee     string `json:"priority_fee"`
		Nonce           string `json:"nonce"`
		GasLimit        string `json:"gas_limit"`
	}

	// Temporary struct for payload
	var tempPayload struct {
		Coin                *v1.Coin              `json:"coin"`
		Outputs             []*v1.Output          `json:"outputs"`
		EthereumSpecific    *tempEthereumSpecific `json:"ethereum_specific"`
		Memo                string                `json:"memo"`
		VaultPublicKeyEcdsa string                `json:"vault_public_key_ecdsa"`
		VaultLocalPartyId   string                `json:"vault_local_party_id"`
	}

	if err := json.Unmarshal(temp.Payload, &tempPayload); err != nil {
		return err
	}

	// Create the proper payload structure
	payload := &v1.KeysignPayload{
		Coin:                tempPayload.Coin,
		Outputs:             tempPayload.Outputs,
		Memo:                &tempPayload.Memo,
		VaultPublicKeyEcdsa: tempPayload.VaultPublicKeyEcdsa,
		VaultLocalPartyId:   tempPayload.VaultLocalPartyId,
	}

	// If ethereum_specific exists, convert and set it
	if tempPayload.EthereumSpecific != nil {
		nonce, err := parseNonce(tempPayload.EthereumSpecific.Nonce)
		if err != nil {
			return fmt.Errorf("failed to parse nonce: %w", err)
		}

		payload.BlockchainSpecific = &v1.KeysignPayload_EthereumSpecific{
			EthereumSpecific: &v1.EthereumSpecific{
				MaxFeePerGasWei: tempPayload.EthereumSpecific.MaxFeePerGasWei,
				PriorityFee:     tempPayload.EthereumSpecific.PriorityFee,
				GasLimit:        tempPayload.EthereumSpecific.GasLimit,
				Nonce:           nonce,
			},
		}
	}

	// Assign values to the KeysignRequest
	k.PublicKey = temp.PublicKey
	k.Payload = payload
	k.SessionID = temp.SessionID
	k.HexEncryptionKey = temp.HexEncryptionKey
	k.DerivePath = temp.DerivePath
	k.IsECDSA = temp.IsECDSA
	k.VaultPassword = temp.VaultPassword
	k.StartSession = temp.StartSession
	k.Parties = temp.Parties

	return nil
}
