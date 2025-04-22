package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/eager7/dogd/btcec"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/vultisig/mobile-tss-lib/tss"
	"github.com/vultisig/vultiserver-plugin/common"
	"github.com/vultisig/vultiserver-plugin/config"
	"github.com/vultisig/vultiserver-plugin/internal/tasks"
	"github.com/vultisig/vultiserver-plugin/internal/types"
)

var isEcdsa bool

// Usage:
//   - start local vultisig services
//   - `go run ./scripts/dev/workers_sign/main.go -ecdsa=true`
func main() {
	flag.BoolVar(&isEcdsa, "ecdsa", true, "user name")
	flag.Parse()

	// policy
	publicKeyEcdsa := "020cdce195caec8f13caa4e807c6c65d1f87d23e65ed4d47f24b137939f9000985"
	publicKeyEddsa := "cbe9f33d38054defe561b9189f0f78721bb4ba836678030ca9db319a11a590ac"
	derivePath := "m/44'/60'/0'/0/0"
	// vault
	hexEncryptionKey := "539440138236b389cb0355aa1e81d11e51e9ad7c94b09bb45704635913604a73" // see plugin/plugin/dca.go
	vaultPassword := "888717"                                                              // see plugin/plugin/dca.go

	// tx hash
	signMessage := []byte("The road to success is always under construction.")

	publicKey := publicKeyEddsa
	if isEcdsa {
		publicKey = publicKeyEcdsa
	}

	keysignRequest := types.KeysignRequest{
		PublicKey:        publicKey,
		Messages:         []string{hex.EncodeToString(signMessage)},
		SessionID:        uuid.New().String(),
		HexEncryptionKey: hexEncryptionKey,
		DerivePath:       derivePath,
		IsECDSA:          isEcdsa,
		VaultPassword:    vaultPassword,
		Parties:          []string{common.PluginPartyID, common.VerifierPartyID},
	}

	err := initiateTxSignWithVerifier(keysignRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	result, err := initiateTxSignWithPlugin(keysignRequest)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Sign result: %v\n", string(result))

	signature, err := parseSignResultIntoRawSignature(result)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to parse signature: %w", err))
		return
	}

	isVerified, err := verifyRawSignature(publicKey, hexEncryptionKey, derivePath, signMessage, signature)
	if err != nil {
		fmt.Println(fmt.Errorf("failed to verify signature: %w", err))
		return
	}

	fmt.Printf("Signature verified: %t\n", isVerified)
}

func initiateTxSignWithPlugin(keysignRequest types.KeysignRequest) ([]byte, error) {
	cfgPlugin, err := config.ReadConfig("config-plugin")
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("failed to read plugin config")
	}

	redisPluginOptions := asynq.RedisClientOpt{
		Addr:     "localhost" + ":" + cfgPlugin.Redis.Port,
		Username: cfgPlugin.Redis.User,
		Password: cfgPlugin.Redis.Password,
		DB:       cfgPlugin.Redis.DB,
	}
	queueClient := asynq.NewClient(redisPluginOptions)
	queueInspector := asynq.NewInspector(redisPluginOptions)

	buf, err := json.Marshal(keysignRequest)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("Failed to marshal local sign request")
	}

	ti, err := queueClient.Enqueue(
		asynq.NewTask(tasks.TypeKeySign, buf),
		asynq.MaxRetry(0),
		asynq.Timeout(2*time.Minute),
		asynq.Retention(5*time.Minute),
		asynq.Queue(tasks.QUEUE_NAME),
	)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("Failed to enqueue signing task")
	}

	fmt.Printf("Enqueued signing task from plugin: %s\n", ti.ID)

	result, err := waitForTaskResult(queueInspector, ti.ID, 120*time.Second)
	if err != nil {
		fmt.Println(err)
		return nil, errors.New("Failed to receive result")
	}

	return result, nil
}

func initiateTxSignWithVerifier(keysignRequest types.KeysignRequest) error {
	cfgVerifier, err := config.ReadConfig("config-verifier")
	if err != nil {
		fmt.Println(err)
		return errors.New("failed to read verifier config")
	}

	redisVerifierOptions := asynq.RedisClientOpt{
		Addr:     "localhost" + ":" + cfgVerifier.Redis.Port,
		Username: cfgVerifier.Redis.User,
		Password: cfgVerifier.Redis.Password,
		DB:       cfgVerifier.Redis.DB,
	}
	queueClient := asynq.NewClient(redisVerifierOptions)

	buf, err := json.Marshal(keysignRequest)
	if err != nil {
		fmt.Println(err)
		return errors.New("Failed to marshal local sign request")
	}

	ti, err := queueClient.Enqueue(
		asynq.NewTask(tasks.TypeKeySign, buf),
		asynq.MaxRetry(0),
		asynq.Timeout(2*time.Minute),
		asynq.Retention(5*time.Minute),
		asynq.Queue(tasks.QUEUE_NAME),
	)
	if err != nil {
		fmt.Println(err)
		return errors.New("Failed to enqueue signing task")
	}

	fmt.Printf("Enqueued signing task from verifier: %s\n", ti.ID)

	return nil
}

func waitForTaskResult(
	queueInspector *asynq.Inspector,
	taskID string,
	timeout time.Duration,
) ([]byte, error) {
	start := time.Now()
	pollInterval := time.Second

	for {
		if time.Since(start) > timeout {
			return nil, fmt.Errorf("timeout waiting for task result after %v", timeout)
		}

		task, err := queueInspector.GetTaskInfo(tasks.QUEUE_NAME, taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task info: %w", err)
		}

		switch task.State {
		case asynq.TaskStateCompleted:
			fmt.Println("Task completed successfully")
			return task.Result, nil
		case asynq.TaskStateArchived:
			return nil, fmt.Errorf("task archived: %s", task.LastErr)
		case asynq.TaskStateRetry:
			fmt.Println("Task scheduled for retry...")
		case asynq.TaskStatePending, asynq.TaskStateActive, asynq.TaskStateScheduled:
			fmt.Println("Task still in progress, waiting...")
		case asynq.TaskStateAggregating:
			fmt.Println("Task aggregating, waiting...")
		default:
			return nil, fmt.Errorf("unexpected task state: %s", task.State)
		}

		time.Sleep(pollInterval)
	}
}

func parseSignResultIntoRawSignature(signResult []byte) ([]byte, error) {
	var keysignResponses map[string]tss.KeysignResponse
	if err := json.Unmarshal(signResult, &keysignResponses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sign result: %w", err)
	}
	var keysignResponse tss.KeysignResponse
	// take the first (and only)
	for _, val := range keysignResponses {
		keysignResponse = val
		break
	}

	r, ok := new(big.Int).SetString(keysignResponse.R, 16)
	if !ok {
		return nil, fmt.Errorf("failed to parse R")
	}

	s, ok := new(big.Int).SetString(keysignResponse.S, 16)
	if !ok {
		return nil, fmt.Errorf("failed to parse S")
	}

	recoveryID := uint8(0)
	if keysignResponse.RecoveryID != "" { // "" for eddsa
		recID, err := strconv.ParseInt(keysignResponse.RecoveryID, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("failed to parse recovery ID: %w", err)
		}
		recoveryID = uint8(recID) // 00 or 01 for ecdsa
	}

	return rawSignature(r, s, recoveryID), nil
}

func verifyRawSignature(vaultPublicKey string, chainCodeHex string, derivePath string, messageHex []byte, signature []byte) (bool, error) {

	// TODO: eddca support, last param is the flag
	derivedPubKeyHex, err := tss.GetDerivedPubKey(strings.TrimPrefix(vaultPublicKey, "0x"), chainCodeHex, derivePath, false)
	if err != nil {
		return false, err
	}

	publicKeyBytes, err := hex.DecodeString(derivedPubKeyHex)
	if err != nil {
		return false, err
	}

	pk, err := btcec.ParsePubKey(publicKeyBytes, btcec.S256())
	if err != nil {
		return false, err
	}

	ecdsaPubKey := ecdsa.PublicKey{
		Curve: btcec.S256(),
		X:     pk.X,
		Y:     pk.Y,
	}
	R := new(big.Int).SetBytes(signature[:32])
	S := new(big.Int).SetBytes(signature[32:64])

	return ecdsa.Verify(&ecdsaPubKey, messageHex, R, S), nil
}

func rawSignature(r *big.Int, s *big.Int, recoveryID uint8) []byte {
	var signature [65]byte
	copy(signature[0:32], r.Bytes())
	copy(signature[32:64], s.Bytes())
	signature[64] = byte(recoveryID)
	return signature[:]
}
