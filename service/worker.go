package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"

	"github.com/vultisig/mobile-tss-lib/tss"

	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/contexthelper"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/plugin"
	"github.com/vultisig/vultisigner/plugin/payroll"
	"github.com/vultisig/vultisigner/relay"
	"github.com/vultisig/vultisigner/storage"
	"github.com/vultisig/vultisigner/storage/postgres"
)

type WorkerService struct {
	cfg          config.Config
	redis        *storage.RedisStorage
	logger       *logrus.Logger
	queueClient  *asynq.Client
	sdClient     *statsd.Client
	blockStorage *storage.BlockStorage
	inspector    *asynq.Inspector
	plugin       plugin.Plugin
	db           storage.DatabaseStorage
	rpcClient    *ethclient.Client
}

// NewWorker creates a new worker service
func NewWorker(cfg config.Config, queueClient *asynq.Client, sdClient *statsd.Client, blockStorage *storage.BlockStorage, inspector *asynq.Inspector) (*WorkerService, error) {
	redis, err := storage.NewRedisStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("storage.NewRedisStorage failed: %w", err)
	}

	db, err := postgres.NewPostgresBackend(false, cfg.Database.DSN)
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}

	rpcClient, err := ethclient.Dial(cfg.RPC.URL)
	if err != nil {
		logrus.Fatalf("Failed to connect to RPC: %v", err)
	}
	var plugin plugin.Plugin

	if cfg.Server.Mode == "pluginserver" {

		switch cfg.Plugin.Type {
		case "payroll":
			plugin = payroll.NewPayrollPlugin(db, logrus.WithField("service", "plugin").Logger, rpcClient)
		default:
			logrus.Fatalf("Invalid plugin type: %s", cfg.Plugin.Type)
		}
	}

	return &WorkerService{
		redis:        redis,
		cfg:          cfg,
		logger:       logrus.WithField("service", "worker").Logger,
		queueClient:  queueClient,
		sdClient:     sdClient,
		blockStorage: blockStorage,
		plugin:       plugin,
		db:           db,
		inspector:    inspector,
		rpcClient:    rpcClient,
	}, nil
}

type KeyGenerationTaskResult struct {
	EDDSAPublicKey string
	ECDSAPublicKey string
}

func (s *WorkerService) incCounter(name string, tags []string) {
	if err := s.sdClient.Count(name, 1, tags, 1); err != nil {
		s.logger.Errorf("fail to count metric, err: %v", err)
	}
}
func (s *WorkerService) measureTime(name string, start time.Time, tags []string) {
	if err := s.sdClient.Timing(name, time.Since(start), tags, 1); err != nil {
		s.logger.Errorf("fail to measure time metric, err: %v", err)
	}
}
func (s *WorkerService) HandleKeyGeneration(ctx context.Context, t *asynq.Task) error {
	if err := contexthelper.CheckCancellation(ctx); err != nil {
		return err
	}
	defer s.measureTime("worker.vault.create.latency", time.Now(), []string{})
	var req types.VaultCreateRequest
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	s.logger.WithFields(logrus.Fields{
		"name":           req.Name,
		"session":        req.SessionID,
		"local_party_id": req.LocalPartyId,
		"email":          req.Email,
	}).Info("Joining keygen")
	s.incCounter("worker.vault.create", []string{})
	if err := req.IsValid(); err != nil {
		return fmt.Errorf("invalid vault create request: %s: %w", err, asynq.SkipRetry)
	}
	keyECDSA, keyEDDSA, err := s.JoinKeyGeneration(req)
	if err != nil {
		_ = s.sdClient.Count("worker.vault.create.error", 1, nil, 1)
		s.logger.Errorf("keygen.JoinKeyGeneration failed: %v", err)
		return fmt.Errorf("keygen.JoinKeyGeneration failed: %v: %w", err, asynq.SkipRetry)
	}

	s.logger.WithFields(logrus.Fields{
		"keyECDSA": keyECDSA,
		"keyEDDSA": keyEDDSA,
	}).Info("localPartyID generation completed")

	result := KeyGenerationTaskResult{
		EDDSAPublicKey: keyEDDSA,
		ECDSAPublicKey: keyECDSA,
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		s.logger.Errorf("json.Marshal failed: %v", err)
		return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if _, err := t.ResultWriter().Write(resultBytes); err != nil {
		s.logger.Errorf("t.ResultWriter.Write failed: %v", err)
		return fmt.Errorf("t.ResultWriter.Write failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}

func (s *WorkerService) HandleKeySign(ctx context.Context, t *asynq.Task) error {
	s.logger.Info("Starting HandleKeySign")
	if err := contexthelper.CheckCancellation(ctx); err != nil {
		s.logger.Error("Context cancelled")
		return err
	}
	var p types.KeysignRequest
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		s.logger.Errorf("json.Unmarshal failed: %v", err)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	defer s.measureTime("worker.vault.sign.latency", time.Now(), []string{})
	s.incCounter("worker.vault.sign", []string{})
	s.logger.WithFields(logrus.Fields{
		"PublicKey":  p.PublicKey,
		"session":    p.SessionID,
		"Messages":   p.Messages,
		"DerivePath": p.DerivePath,
		"IsECDSA":    p.IsECDSA,
	}).Info("joining keysign")

	signatures, err := s.JoinKeySign(p)
	if err != nil {
		s.logger.Errorf("join keysign failed: %v", err)
		return fmt.Errorf("join keysign failed: %v: %w", err, asynq.SkipRetry)
	}

	s.logger.WithFields(logrus.Fields{
		"Signatures": signatures,
	}).Info("localPartyID sign completed")

	resultBytes, err := json.Marshal(signatures)
	if err != nil {
		s.logger.Errorf("json.Marshal failed: %v", err)
		return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
	}

	if _, err := t.ResultWriter().Write(resultBytes); err != nil {
		s.logger.Errorf("t.ResultWriter.Write failed: %v", err)
		return fmt.Errorf("t.ResultWriter.Write failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}
func (s *WorkerService) HandleEmailVaultBackup(ctx context.Context, t *asynq.Task) error {
	if err := contexthelper.CheckCancellation(ctx); err != nil {
		return err
	}
	s.incCounter("worker.vault.backup.email", []string{})
	var req types.EmailRequest
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		s.logger.Errorf("json.Unmarshal failed: %v", err)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}
	s.logger.WithFields(logrus.Fields{
		"email":    req.Email,
		"filename": req.FileName,
	}).Info("sending email")
	emailServer := "https://mandrillapp.com/api/1.0/messages/send-template"
	payload := MandrillPayload{
		Key:          s.cfg.EmailServer.ApiKey,
		TemplateName: "fastvault",
		TemplateContent: []MandrilMergeVarContent{
			{
				Name:    "VAULT_NAME",
				Content: req.VaultName,
			},
			{
				Name:    "VERIFICATION_CODE",
				Content: req.Code,
			},
		},
		Message: MandrillMessage{
			To: []MandrillTo{
				{
					Email: req.Email,
					Type:  "to",
				},
			},
			MergeVars: []MandrillVar{
				{
					Rcpt: req.Email,
					Vars: []MandrilMergeVarContent{
						{
							Name:    "VAULT_NAME",
							Content: req.VaultName,
						},
						{
							Name:    "VERIFICATION_CODE",
							Content: req.Code,
						},
					},
				},
			},
			SendingDomain: "vultisig.com",
			Attachments: []MandrillAttachment{
				{
					Type:    "application/octet-stream",
					Name:    req.FileName,
					Content: base64.StdEncoding.EncodeToString([]byte(req.FileContent)),
				},
			},
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		s.logger.Errorf("json.Marshal failed: %v", err)
		return fmt.Errorf("json.Marshal failed: %v: %w", err, asynq.SkipRetry)
	}
	resp, err := http.Post(emailServer, "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		s.logger.Errorf("http.Post failed: %v", err)
		return fmt.Errorf("http.Post failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.logger.Errorf("failed to close body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		s.logger.Errorf("http.Post failed: %s", resp.Status)
		return fmt.Errorf("http.Post failed: %s: %w", resp.Status, asynq.SkipRetry)
	}
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Errorf("io.ReadAll failed: %v", err)
		return fmt.Errorf("io.ReadAll failed: %w", err)
	}
	s.logger.Info(string(result))
	if _, err := t.ResultWriter().Write([]byte("email sent")); err != nil {
		return fmt.Errorf("t.ResultWriter.Write failed: %v", err)
	}
	return nil
}

func (s *WorkerService) HandleReshare(ctx context.Context, t *asynq.Task) error {
	if err := contexthelper.CheckCancellation(ctx); err != nil {
		return err
	}
	var req types.ReshareRequest
	if err := json.Unmarshal(t.Payload(), &req); err != nil {
		s.logger.Errorf("json.Unmarshal failed: %v", err)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	defer s.measureTime("worker.vault.reshare.latency", time.Now(), []string{})
	s.incCounter("worker.vault.reshare", []string{})
	s.logger.WithFields(logrus.Fields{
		"name":           req.Name,
		"session":        req.SessionID,
		"local_party_id": req.LocalPartyId,
		"email":          req.Email,
	}).Info("reshare request")
	if err := req.IsValid(); err != nil {
		return fmt.Errorf("invalid reshare request: %s: %w", err, asynq.SkipRetry)
	}
	localState, err := relay.NewLocalStateAccessorImp(req.LocalPartyId, s.cfg.Server.VaultsFilePath, req.PublicKey, req.EncryptionPassword, s.blockStorage)
	if err != nil {
		s.logger.Errorf("relay.NewLocalStateAccessorImp failed: %v", err)
		return fmt.Errorf("relay.NewLocalStateAccessorImp failed: %v: %w", err, asynq.SkipRetry)
	}
	var vault *vaultType.Vault
	if localState.Vault != nil {
		// reshare vault
		vault = localState.Vault

	} else {
		vault = &vaultType.Vault{
			Name:           req.Name,
			PublicKeyEcdsa: "",
			PublicKeyEddsa: "",
			HexChainCode:   req.HexChainCode,
			LocalPartyId:   req.LocalPartyId,
			Signers:        req.OldParties,
			ResharePrefix:  req.OldResharePrefix,
		}
		// create new vault
	}
	if err := s.Reshare(vault,
		req.SessionID,
		req.HexEncryptionKey,
		s.cfg.Relay.Server,
		req.EncryptionPassword,
		req.Email); err != nil {
		s.logger.Errorf("reshare failed: %v", err)
		return fmt.Errorf("reshare failed: %v: %w", err, asynq.SkipRetry)
	}

	return nil
}

func (s *WorkerService) HandlePluginTransaction(ctx context.Context, t *asynq.Task) error {
	s.logger.Info("Starting HandlePluginTransaction")

	if err := contexthelper.CheckCancellation(ctx); err != nil {
		return err
	}

	var triggerEvent types.PluginTriggerEvent
	if err := json.Unmarshal(t.Payload(), &triggerEvent); err != nil {
		s.logger.Errorf("json.Unmarshal failed: %v", err)
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	defer s.measureTime("worker.plugin.transaction.latency", time.Now(), []string{})
	s.incCounter("worker.plugin.transaction", []string{})

	policy, err := s.db.GetPluginPolicy(triggerEvent.PolicyID)
	if err != nil {
		s.logger.Errorf("db.GetPluginPolicy failed: %v", err)
		return fmt.Errorf("db.GetPluginPolicy failed: %v: %w", err, asynq.SkipRetry)
	}

	s.logger.WithFields(logrus.Fields{
		"policy_id":         policy.ID,
		"public_key":        policy.PublicKey,
		"plugin_type":       policy.PluginType,
		"policy public key": policy.PublicKey,
	}).Info("Retrieved policy for signing")

	signRequests, err := s.plugin.ProposeTransactions(policy)
	if err != nil {
		s.logger.Errorf("Failed to create signing request: %v", err)
		return fmt.Errorf("failed to create signing request: %v: %w", err, asynq.SkipRetry)
	}

	for _, signRequest := range signRequests {

		policyUUID, err := uuid.Parse(signRequest.PolicyID)
		if err != nil {
			s.logger.Errorf("Failed to parse policy ID as UUID: %v", err)
			return err
		}

		// create transaction with PENDING status
		metadata := map[string]interface{}{
			"timestamp":  time.Now(),
			"plugin_id":  signRequest.PluginID,
			"public_key": signRequest.KeysignRequest.PublicKey,
		}

		newTx := types.TransactionHistory{
			PolicyID: policyUUID,
			TxBody:   signRequest.Transaction,
			Status:   types.StatusPending,
			Metadata: metadata,
		}

		txID, err := s.db.CreateTransactionHistory(newTx) //where to store txId? what is the best way to retrieve a tx?	Maybe just keep it in this context, and if status is failed at the end then we drop this instance and restart a new one later?
		if err != nil {
			s.logger.Errorf("Failed to create transaction history: %v", err)
			continue
		}

		// start TSS signing process
		signBytes, err := json.Marshal(signRequest)
		if err != nil {
			s.logger.Errorf("Failed to marshal sign request: %v", err)
			continue
		}

		signResp, err := http.Post(
			fmt.Sprintf("http://localhost:%d/signFromPlugin", 8080),
			"application/json",
			bytes.NewBuffer(signBytes),
		)
		if err != nil {
			metadata["error"] = err.Error()
			s.db.UpdateTransactionStatus(txID, types.StatusSigningFailed, metadata)
			s.logger.Errorf("Failed to make sign request: %v", err)
			return err
		}
		defer signResp.Body.Close()

		respBody, err := io.ReadAll(signResp.Body)
		if err != nil {
			s.logger.Errorf("Failed to read response: %v", err)
			return err
		}

		if signResp.StatusCode != http.StatusOK {
			metadata["error"] = string(respBody)
			s.db.UpdateTransactionStatus(txID, types.StatusSigningFailed, metadata)
			s.logger.Errorf("Failed to sign transaction: %s", string(respBody))
			return fmt.Errorf("failed to sign transaction: %s", string(respBody))
		}

		// prepare local sign request
		signRequest.KeysignRequest.StartSession = true
		signRequest.KeysignRequest.Parties = []string{"1", "2"}
		buf, err := json.Marshal(signRequest.KeysignRequest)
		if err != nil {
			s.logger.Errorf("Failed to marshal local sign request: %v", err)
			return err
		}

		// Enqueue TypeKeySign directly
		ti, err := s.queueClient.Enqueue(
			asynq.NewTask(tasks.TypeKeySign, buf),
			asynq.MaxRetry(0),
			asynq.Timeout(2*time.Minute),
			asynq.Retention(5*time.Minute),
			asynq.Queue(tasks.QUEUE_NAME),
		)
		if err != nil {
			s.logger.Errorf("Failed to enqueue signing task: %v", err)
			continue
		}

		s.logger.Infof("Enqueued signing task: %s", ti.ID)

		// wait for result with timeout
		result, err := s.waitForTaskResult(ti.ID, 120*time.Second) // adjust timeout as needed (each policy provider should be able to set it, but there should be an incentive to not retry too much)
		if err != nil {                                            //do we consider that the signature is always valid if err = nil?
			metadata["error"] = err.Error()
			metadata["task_id"] = ti.ID
			s.db.UpdateTransactionStatus(txID, types.StatusSigningFailed, metadata)
			s.logger.Errorf("Failed to get task result: %v", err)
			return err
		}

		// Update to SIGNED status with result
		metadata["task_id"] = ti.ID
		metadata["result"] = result
		if err := s.db.UpdateTransactionStatus(txID, types.StatusSigned, metadata); err != nil {
			s.logger.Errorf("Failed to update transaction status: %v", err)
		}

		if err := s.db.UpdateTriggerExecution(policy.ID); err != nil { //todo : check why this seems to work even when tss fails
			s.logger.Errorf("Failed to update last execution: %v", err)
		}

		s.logger.Infof("Plugin signing test complete. Status: %d, Response: %s",
			signResp.StatusCode, string(respBody))

		//todo : retry

		///////////////////////////////////////////////////////////////////////

		var signatures map[string]tss.KeysignResponse
		if err := json.Unmarshal(result, &signatures); err != nil {
			s.logger.Errorf("Failed to unmarshal signatures: %v", err)
			return fmt.Errorf("failed to unmarshal signatures: %w", err)
		}

		var signature tss.KeysignResponse
		for _, sig := range signatures {
			signature = sig
			break
		}

		r, s_value, originalTx, chainID, recoveryID, v, err := s.convertRAndS(signature, signRequest)
		if err != nil {
			s.logger.Errorf("Failed to convert R and S: %v", err)
			return fmt.Errorf("failed to convert R and S: %w", err)
		}

		///////////

		ethAddress21, ethAddress22, err := s.convertPolicyPublicKeyToAddress2(policy)
		if err != nil {
			s.logger.Errorf("Failed to convert policy public key to address: %v", err)
			return fmt.Errorf("failed to convert policy public key to address: %w", err)
		}

		innerTx := &gtypes.LegacyTx{
			Nonce:    originalTx.Nonce(),
			GasPrice: originalTx.GasPrice(),
			Gas:      originalTx.Gas(),
			To:       originalTx.To(),
			Value:    originalTx.Value(),
			Data:     originalTx.Data(),
			V:        v,
			R:        r,
			S:        s_value,
		}

		signedTx := gtypes.NewTx(innerTx)
		signer := gtypes.NewLondonSigner(chainID)
		sender, err := signer.Sender(signedTx)
		if err != nil {
			s.logger.Errorf("Failed to get sender: %v", err)
			return fmt.Errorf("failed to get sender: %w", err)
		}

		fmt.Printf("Signature from MPC \n")
		fmt.Printf("r (as string): %s\n", signature.R) // print as string
		fmt.Printf("r (as hex): %x\n", signature.R)    // print as hex
		fmt.Printf("r (type): %T\n", signature.R)

		fmt.Printf("s (as string): %s\n", signature.S)
		fmt.Printf("s (as hex): %x\n", signature.S)
		fmt.Printf("s (type): %T\n", signature.S)

		fmt.Printf("recovery_id (as string): %s\n", signature.RecoveryID)
		fmt.Printf("recovery_id (as hex): %x\n", signature.RecoveryID)
		fmt.Printf("recovery_id (type): %T\n", signature.RecoveryID)

		fmt.Printf("Hash sent to MPC and hash used to sign \n")
		rawTx, _ := rlp.EncodeToBytes(originalTx) //from  signrequest.Transaction
		txHash := crypto.Keccak256(rawTx)
		messageBytes, err := hex.DecodeString(string(signRequest.Messages[0])) //from signRequest.message[0]
		if err != nil {
			s.logger.WithError(err).Error("Failed to decode message hex")
		}
		fmt.Printf("hash used to sign : %x\n", txHash)
		fmt.Printf("hash_sent_to_mpc (decoded): %x\n", messageBytes)
		fmt.Printf("Hashes match( from signRequest.Messages[0] and from signRequest.Transaction) : %v\n", bytes.Equal(messageBytes, txHash))

		fmt.Printf("Comparing recovery methods \n")
		fmt.Printf("Public key from policy: %s\n", policy.PublicKey)
		fmt.Printf("Eth address derived from policy public key: %s\n", ethAddress21.Hex())
		fmt.Printf("Eth address derived from policy public key: %s\n", ethAddress22.Hex())

		fmt.Printf("Sender from signer: %s\n", sender.Hex())
		fmt.Printf("Matches: %t\n", ethAddress21 == sender)
		fmt.Printf("Matches: %t\n", ethAddress22 == sender)

		s.testSignatureRecovery(originalTx.Hash().Bytes(), r, s_value, recoveryID, chainID)

		s.logger.WithField("rpcClient", s.rpcClient).Info("Attempting to send transaction")
		err = s.rpcClient.SendTransaction(ctx, signedTx)
		if err != nil {
			s.logger.WithField("rpcClient", s.rpcClient).WithError(err).Error("Failed to send transaction")
			s.logger.WithError(err).Error("Failed to send transaction")
			return fmt.Errorf("failed to send transaction: %w", err)
		}

		fmt.Printf("Transaction sent successfully \n")
		fmt.Printf("txHash: %s\n", signedTx.Hash().Hex())
		fmt.Printf("sender: %s\n", sender.Hex())

		///
		/*ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) //how much time should we monitor the tx?
		defer cancel()

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		txHash := signedTx.Hash()
		for {
			select {
			case <-ctx.Done():
				s.logger.WithField("txHash", signedTx.Hash().Hex()).Info("Transaction monitoring timed out")
				return &types.TransactionError{
					Code:    types.ErrTxTimeout,
					Message: fmt.Sprintf("Transaction monitoring timed out for tx: %s", txHash.Hex()),
				}

			case <-ticker.C:
				// check tx status
				_, isPending, err := s.rpcClient.TransactionByHash(ctx, txHash)
				if err != nil {
					if err == ethereum.NotFound {
						s.logger.WithField("txHash", signedTx.Hash().Hex()).Info("Transaction dropped from mempool")
						return &types.TransactionError{
							Code:    types.ErrTxDropped,
							Message: fmt.Sprintf("Transaction dropped from mempool: %s", txHash.Hex()),
						}
					}
					continue // keep trying on other RPC errors
				}

				if !isPending {
					receipt, err := s.rpcClient.TransactionReceipt(ctx, txHash)
					if err != nil {
						s.logger.WithField("txHash", signedTx.Hash().Hex()).Errorf("Failed to get transaction receipt: %v", err)
						continue
					}

					if receipt.Status == 0 {
						// try to get revert reason
						//reason := s.plugin.getRevertReason(ctx, signedTx, receipt.BlockNumber)
						//return &types.TransactionError{
						//	Code:    types.ErrExecutionReverted,
						//	Message: fmt.Sprintf("Transaction reverted: %s", reason),
						//}
					}

					// Transaction successful
					return nil
				}
			}
		}
		///////////////////////////////////////////////////////////////////////

		s.logger.Info("About to call SigningComplete")
		signingComplete := s.plugin.SigningComplete( //todo : remove occurances of SignedTransaction type if not neede
			signedTx,
		)
		s.logger.Info("After SigningComplete call")

		if signingComplete != nil {
			s.logger.WithFields(logrus.Fields{
				"error":      signingComplete,
				"chainID":    chainID,
				"v":          v,
				"r":          r,
				"s":          s_value,
				"recoveryID": recoveryID,
			}).Error("Failed to sign transaction TEST")
			return signingComplete
		}*/
	}

	return nil
}

func (s *WorkerService) convertRAndS(signature tss.KeysignResponse, signRequest types.PluginKeysignRequest) (r *big.Int, s_value *big.Int, originalTx *gtypes.Transaction, chainID *big.Int, recoveryID int64, v *big.Int, err error) {
	// convert R and S from hex strings to big.Int
	r = new(big.Int)
	r.SetString(signature.R, 16) // base 16 for hex
	if r == nil {
		return nil, nil, nil, nil, 0, nil, fmt.Errorf("failed to parse R value")
	}

	s_value = new(big.Int)
	s_value.SetString(signature.S, 16) // base 16 for hex
	if s_value == nil {
		return nil, nil, nil, nil, 0, nil, fmt.Errorf("failed to parse S value")
	}

	txBytes, err := hex.DecodeString(signRequest.Transaction)
	if err != nil {
		s.logger.Errorf("Failed to decode transaction hex: %v", err)
		return nil, nil, nil, nil, 0, nil, fmt.Errorf("failed to decode transaction hex: %w", err)
	}

	originalTx = new(gtypes.Transaction)
	if err := originalTx.UnmarshalBinary(txBytes); err != nil {
		s.logger.Errorf("Failed to unmarshal transaction: %v", err)
		return nil, nil, nil, nil, 0, nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	chainID = big.NewInt(137) // polygon mainnet chain ID
	// calculate V according to EIP-155
	recoveryID, err = strconv.ParseInt(signature.RecoveryID, 10, 64)
	if err != nil {
		return nil, nil, nil, nil, 0, nil, fmt.Errorf("failed to parse recovery ID: %w", err)
	}

	v = new(big.Int).Set(chainID)
	v.Mul(v, big.NewInt(2))
	v.Add(v, big.NewInt(35+recoveryID))

	return r, s_value, originalTx, chainID, recoveryID, v, nil
}

func (s *WorkerService) waitForTaskResult(taskID string, timeout time.Duration) ([]byte, error) {
	start := time.Now()
	pollInterval := time.Second

	for {
		if time.Since(start) > timeout {
			return nil, fmt.Errorf("timeout waiting for task result after %v", timeout)
		}

		task, err := s.inspector.GetTaskInfo(tasks.QUEUE_NAME, taskID)
		if err != nil {
			return nil, fmt.Errorf("failed to get task info: %w", err)
		}

		switch task.State {
		case asynq.TaskStateCompleted:
			s.logger.Info("Task completed successfully")
			return task.Result, nil
		case asynq.TaskStateArchived:
			return nil, fmt.Errorf("task archived: %s", task.LastErr)
		case asynq.TaskStateRetry:
			s.logger.Debug("Task scheduled for retry...")
		case asynq.TaskStatePending, asynq.TaskStateActive, asynq.TaskStateScheduled:
			s.logger.Debug("Task still in progress, waiting...")
		case asynq.TaskStateAggregating:
			s.logger.Debug("Task aggregating, waiting...")
		default:
			return nil, fmt.Errorf("unexpected task state: %s", task.State)
		}

		time.Sleep(pollInterval)
	}
}

func (s *WorkerService) testSignatureRecovery(txHash []byte, r, s_value *big.Int, recoveryID int64, chainID *big.Int) {
	fmt.Printf("\nRecovering Address from Signature:\n")

	sig := make([]byte, 65)
	copy(sig[:32], r.Bytes())
	copy(sig[32:64], s_value.Bytes())
	sig[64] = byte(recoveryID)

	pubKeyBytes, err := crypto.Ecrecover(txHash, sig)
	if err != nil {
		s.logger.WithError(err).Error("Failed to recover public key")
		return
	}
	fmt.Printf("1. Recovered public key (hex): %x\n", pubKeyBytes)
	fmt.Printf("   Prefix: %x\n", pubKeyBytes[0])
	fmt.Printf("   X: %x\n", pubKeyBytes[1:33])
	fmt.Printf("   Y: %x\n", pubKeyBytes[33:])

	pubKey, err := crypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		s.logger.WithError(err).Error("Failed to unmarshal public key")
		return
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	fmt.Printf("2. Final address: %s\n", recoveredAddr.Hex())
}

func (s *WorkerService) convertPolicyPublicKeyToAddress2(policy types.PluginPolicy) (common.Address, common.Address, error) {
	fmt.Printf("\nDEBUG: Policy Public Key Conversion\n")
	fmt.Printf(" Input public key: %s\n", policy.PublicKey)

	publicKeyBytes, err := hex.DecodeString(policy.PublicKey)
	if err != nil {
		return common.Address{}, common.Address{}, fmt.Errorf("failed to decode public key: %w", err)
	}
	fmt.Printf(" After hex decode (%d bytes):\n", len(publicKeyBytes))
	fmt.Printf("   Full bytes: %x\n", publicKeyBytes)
	fmt.Printf("   First byte: %x (should be 02 or 03)\n", publicKeyBytes[0])
	fmt.Printf("   Rest: %x\n", publicKeyBytes[1:])

	pubKey, err := crypto.DecompressPubkey(publicKeyBytes)
	if err != nil {
		return common.Address{}, common.Address{}, fmt.Errorf("failed to decompress public key: %w", err)
	}

	uncompressedBytes := crypto.FromECDSAPub(pubKey)
	fmt.Printf(" After conversion to uncompressed:\n")
	fmt.Printf("   Full bytes: %x\n", uncompressedBytes)
	fmt.Printf("   Prefix: %x\n", uncompressedBytes[0])
	fmt.Printf("   X: %x\n", uncompressedBytes[1:33])
	fmt.Printf("   Y: %x\n", uncompressedBytes[33:])

	pubKey2, err := crypto.UnmarshalPubkey(uncompressedBytes)
	if err != nil {
		return common.Address{}, common.Address{}, fmt.Errorf("failed to unmarshal uncompressed key: %w", err)
	}

	addr1 := crypto.PubkeyToAddress(*pubKey2)

	/*uncompressed1 := crypto.FromECDSAPub(pubKey)
	hash1 := crypto.Keccak256(uncompressed1[1:])
	var addr1 common.Address
	copy(addr1[:], hash1[12:])
	fmt.Printf("3a. First possible address: %s\n", addr1.Hex())*/

	return addr1, addr1, nil
}
