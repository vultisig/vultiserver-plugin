package sigutil

import (
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/vultisig/mobile-tss-lib/tss"
	"math/big"
	"testing"
)

func TestSignLegacyTx(t *testing.T) {
	// Setup common test data

	privKey, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(privKey.PublicKey)
	chainID := big.NewInt(1)

	// Create a sample unsigned transaction
	unsignedTx := types.NewTransaction(
		0,
		common.HexToAddress("0xRecipient"),
		big.NewInt(100),
		21000,
		big.NewInt(1e9),
		[]byte{})

	rawTx, _ := unsignedTx.MarshalBinary()

	signer := types.NewEIP155Signer(chainID)
	txHash := signer.Hash(unsignedTx).Bytes()
	sig, _ := crypto.Sign(txHash, privKey)

	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:64])
	recoveryID := sig[64]

	tests := []struct {
		name       string
		keysignR   string
		keysignS   string
		recoveryID string
		rawTx      string
		chainID    *big.Int
		wantErr    bool
		errorMsg   string
		validate   func(t *testing.T, tx *types.Transaction, sender *common.Address)
	}{
		{
			name:       "valid signature",
			keysignR:   fmt.Sprintf("%x", r),
			keysignS:   fmt.Sprintf("%x", s),
			recoveryID: fmt.Sprintf("%d", recoveryID),
			rawTx:      hex.EncodeToString(rawTx),
			chainID:    chainID,
			validate: func(t *testing.T, tx *types.Transaction, sender *common.Address) {
				recoveredSender, err := types.NewEIP155Signer(chainID).Sender(tx)
				require.NoError(t, err)

				require.True(t, unsignedTx.Hash().Hex() != "", "unsigned tx hash is missing")
				require.Equal(t, addr, *sender, "sender address mismatch")
				require.Equal(t, recoveredSender, *sender, "recovered sender address address mismatch")
			},
		},
		{
			name:       "invalid R value",
			keysignR:   "invalid_hex",
			keysignS:   fmt.Sprintf("%x", s),
			recoveryID: fmt.Sprintf("%d", recoveryID),
			rawTx:      hex.EncodeToString(rawTx),
			chainID:    chainID,
			wantErr:    true,
			errorMsg:   "failed to parse R",
		},
		{
			name:       "invalid S value",
			keysignR:   fmt.Sprintf("%x", r),
			keysignS:   "invalid_hex",
			recoveryID: fmt.Sprintf("%d", recoveryID),
			rawTx:      hex.EncodeToString(rawTx),
			chainID:    chainID,
			wantErr:    true,
			errorMsg:   "failed to parse S",
		},
		{
			name:       "invalid recoveryID",
			keysignR:   fmt.Sprintf("%x", r),
			keysignS:   fmt.Sprintf("%x", s),
			recoveryID: "invalid_recovery_id",
			rawTx:      hex.EncodeToString(rawTx),
			chainID:    chainID,
			wantErr:    true,
			errorMsg:   "failed to parse recovery ID",
		},
		{
			name:       "invalid raw transaction",
			keysignR:   fmt.Sprintf("%x", r),
			keysignS:   fmt.Sprintf("%x", s),
			recoveryID: fmt.Sprintf("%d", recoveryID),
			rawTx:      "invalid_hex",
			chainID:    chainID,
			wantErr:    true,
			errorMsg:   "failed to decode raw transaction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			keysignResp := tss.KeysignResponse{
				R:          tt.keysignR,
				S:          tt.keysignS,
				RecoveryID: tt.recoveryID,
			}

			tx, sender, err := SignLegacyTx(keysignResp, "", tt.rawTx, tt.chainID)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errorMsg)
				return
			}
			require.NoError(t, err)
			tt.validate(t, tx, sender)
		})
	}
}

func TestVerifySignature(t *testing.T) {

	privKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	pubKeyBytes := crypto.FromECDSAPub(&privKey.PublicKey)
	pubKeyHex := hex.EncodeToString(pubKeyBytes)

	// Test message and signature
	message := []byte("test message")
	messageHash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)))
	ethSig, err := crypto.Sign(messageHash, privKey)
	require.NoError(t, err)

	signature := ethSig[:64]

	tests := []struct {
		name         string
		vaultPubKey  string
		chainCodeHex string
		derivePath   string
		messageHex   []byte
		signature    []byte
		wantedResult bool
		wantErr      bool
		errorMsg     string
	}{
		{
			name:         "valid signature",
			vaultPubKey:  pubKeyHex,
			chainCodeHex: "539440138236b389cb0355aa1e81d11e51e9ad7c94b09bb45704635913604a73",
			derivePath:   "m/44'/60'/0'/0/0",
			messageHex:   message,
			signature:    signature,
			wantedResult: true,
			wantErr:      false,
		},
		{
			name:         "different message",
			vaultPubKey:  pubKeyHex,
			chainCodeHex: "539440138236b389cb0355aa1e81d11e51e9ad7c94b09bb45704635913604a73",
			derivePath:   "m/44'/60'/0'/0/0",
			messageHex:   []byte("different message"),
			signature:    signature,
			wantedResult: false,
			wantErr:      false,
		},
		{
			name:         "invalid public key",
			vaultPubKey:  "invalid_hex",
			chainCodeHex: "539440138236b389cb0355aa1e81d11e51e9ad7c94b09bb45704635913604a73",
			derivePath:   "m/44'/60'/0'/0/0",
			messageHex:   message,
			signature:    signature,
			wantedResult: false,
			wantErr:      true,
			errorMsg:     "decode hex pub key failed",
		},
		{
			name:         "tampered signature",
			vaultPubKey:  pubKeyHex,
			chainCodeHex: "539440138236b389cb0355aa1e81d11e51e9ad7c94b09bb45704635913604a73",
			derivePath:   "m/44'/60'/0'/0/0",
			messageHex:   message,
			signature:    append(signature[:10], make([]byte, len(signature)-10)...),
			wantedResult: false,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := VerifySignature(tt.vaultPubKey, tt.chainCodeHex, tt.derivePath, tt.messageHex, tt.signature)

			if tt.wantErr {
				require.Error(t, err)
				fmt.Println(err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantedResult, result)
		})
	}

}
