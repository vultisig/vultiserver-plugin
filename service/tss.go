package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	keygen "github.com/vultisig/commondata/go/vultisig/keygen/v1"
	"github.com/vultisig/mobile-tss-lib/tss"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/relay"

	vaultType "github.com/vultisig/commondata/go/vultisig/vault/v1"
)

type VaultOperation interface {
	BackupVault(req types.VaultCreateRequest, partiesJoined []string, ecdsaPubkey, eddsaPubkey, hexChainCode string, localStateAccessor *relay.LocalStateAccessorImp) error
	SaveVaultAndScheduleEmail(vault *vaultType.Vault, encryptionPassword, email string) error
}

func (s *WorkerService) JoinKeyGeneration(req types.VaultCreateRequest) (string, string, error) {
	keyFolder := s.cfg.VaultsFilePath
	serverURL := s.cfg.Relay.Server
	relayClient := relay.NewRelayClient(serverURL)

	if req.StartSession {
		if err := relayClient.StartSession(req.SessionID, req.Parties); err != nil {
			return "", "", fmt.Errorf("failed to start session: %w", err)
		}
	} else {
		// Let's register session here
		if err := relayClient.RegisterSession(req.SessionID, req.LocalPartyId); err != nil {
			return "", "", fmt.Errorf("failed to register session: %w", err)
		}
	}
	// wait longer for keygen start
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	partiesJoined, err := relayClient.WaitForSessionStart(ctx, req.SessionID)
	s.logger.WithFields(logrus.Fields{
		"sessionID":      req.SessionID,
		"parties_joined": partiesJoined,
	}).Info("Session started")

	if err != nil {
		return "", "", fmt.Errorf("failed to wait for session start: %w", err)
	}

	localStateAccessor, err := relay.NewLocalStateAccessorImp(keyFolder, "", "", s.blockStorage)
	if err != nil {
		return "", "", fmt.Errorf("failed to create localStateAccessor: %w", err)
	}

	tssServerImp, err := s.createTSSService(serverURL, req.SessionID, req.HexEncryptionKey, localStateAccessor, true, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to create TSS service: %w", err)
	}

	ecdsaPubkey, eddsaPubkey := "", ""
	endCh, wg := s.startMessageDownload(serverURL, req.SessionID, req.LocalPartyId, req.HexEncryptionKey, tssServerImp, "")
	for attempt := 0; attempt < 3; attempt++ {
		ecdsaPubkey, eddsaPubkey, err = s.keygenWithRetry(req, partiesJoined, tssServerImp)
		if err == nil {
			break
		}
	}
	close(endCh)
	wg.Wait()
	if err != nil {
		return "", "", err
	}

	if err := relayClient.CompleteSession(req.SessionID, req.LocalPartyId); err != nil {
		s.logger.WithFields(logrus.Fields{
			"session": req.SessionID,
			"error":   err,
		}).Error("Failed to complete session")
	}

	if isCompleted, err := relayClient.CheckCompletedParties(req.SessionID, partiesJoined); err != nil || !isCompleted {
		s.logger.WithFields(logrus.Fields{
			"sessionID":   req.SessionID,
			"isCompleted": isCompleted,
			"error":       err,
		}).Error("Failed to check completed parties")
	}

	err = s.BackupVault(req, partiesJoined, ecdsaPubkey, eddsaPubkey, req.HexChainCode, localStateAccessor)
	if err != nil {
		return "", "", fmt.Errorf("failed to backup vault: %w", err)
	}

	return ecdsaPubkey, eddsaPubkey, nil
}

func (s *WorkerService) keygenWithRetry(req types.VaultCreateRequest, partiesJoined []string, tssService tss.Service) (string, string, error) {
	resp, err := s.generateECDSAKey(tssService, req, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate ECDSA key: %w", err)
	}

	respEDDSA, err := s.generateEDDSAKey(tssService, req, partiesJoined)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate EDDSA key: %w", err)
	}
	return resp.PubKey, respEDDSA.PubKey, nil
}

func (s *WorkerService) generateECDSAKey(tssService tss.Service, req types.VaultCreateRequest, partiesJoined []string) (*tss.KeygenResponse, error) {
	defer s.measureTime("worker.vault.create.ECDSA.latency", time.Now(), []string{})
	s.logger.WithFields(logrus.Fields{
		"local_party_id": req.LocalPartyId,
		"chain_code":     req.HexChainCode,
		"parties_joined": partiesJoined,
	}).Info("Start ECDSA keygen...")
	resp, err := tssService.KeygenECDSA(&tss.KeygenRequest{
		LocalPartyID: req.LocalPartyId,
		AllParties:   strings.Join(partiesJoined, ","),
		ChainCodeHex: req.HexChainCode,
	})
	if err != nil {
		return nil, fmt.Errorf("generate ECDSA key: %w", err)
	}
	s.logger.WithFields(logrus.Fields{
		"local_party_id": req.LocalPartyId,
		"pub_key":        resp.PubKey,
	}).Info("ECDSA keygen response")
	time.Sleep(time.Second)
	return resp, nil
}

func (s *WorkerService) generateEDDSAKey(tssService tss.Service, req types.VaultCreateRequest, partiesJoined []string) (*tss.KeygenResponse, error) {
	defer s.measureTime("worker.vault.create.EDDSA.latency", time.Now(), []string{})
	s.logger.WithFields(logrus.Fields{
		"local_party_id": req.LocalPartyId,
		"chain_code":     req.HexChainCode,
		"parties_joined": partiesJoined,
	}).Info("Start EDDSA keygen...")
	resp, err := tssService.KeygenEdDSA(&tss.KeygenRequest{
		LocalPartyID: req.LocalPartyId,
		AllParties:   strings.Join(partiesJoined, ","),
		ChainCodeHex: req.HexChainCode,
	})
	if err != nil {
		return nil, fmt.Errorf("generate EDDSA key: %w", err)
	}
	s.logger.WithFields(logrus.Fields{
		"local_party_id": req.LocalPartyId,
		"pub_key":        resp.PubKey,
	}).Info("EDDSA keygen response")
	time.Sleep(time.Second)
	return resp, nil
}

func (s *WorkerService) BackupVault(req types.VaultCreateRequest,
	partiesJoined []string,
	ecdsaPubkey, eddsaPubkey string,
	hexChainCode string,
	localStateAccessor *relay.LocalStateAccessorImp) error {
	ecdsaKeyShare, err := localStateAccessor.GetLocalState(ecdsaPubkey)
	if err != nil {
		return fmt.Errorf("failed to get local sate: %w", err)
	}

	eddsaKeyShare, err := localStateAccessor.GetLocalState(eddsaPubkey)
	if err != nil {
		return fmt.Errorf("failed to get local sate: %w", err)
	}

	vault := &vaultType.Vault{
		Name:           req.Name,
		PublicKeyEcdsa: ecdsaPubkey,
		PublicKeyEddsa: eddsaPubkey,
		Signers:        partiesJoined,
		CreatedAt:      timestamppb.New(time.Now()),
		HexChainCode:   hexChainCode,
		KeyShares: []*vaultType.Vault_KeyShare{
			{
				PublicKey: ecdsaPubkey,
				Keyshare:  ecdsaKeyShare,
			},
			{
				PublicKey: eddsaPubkey,
				Keyshare:  eddsaKeyShare,
			},
		},
		LocalPartyId:  req.LocalPartyId,
		ResharePrefix: "",
	}
	if req.LibType == types.DKLS {
		vault.LibType = keygen.LibType_LIB_TYPE_DKLS
	} else {
		vault.LibType = keygen.LibType_LIB_TYPE_GG20
	}
	return s.SaveVaultAndScheduleEmail(vault, req.EncryptionPassword, req.Email)
}

func (s *WorkerService) createTSSService(serverURL, Session, HexEncryptionKey string, localStateAccessor tss.LocalStateAccessor, createPreParam bool, messageID string) (*tss.ServiceImpl, error) {
	messenger := relay.NewMessenger(serverURL, Session, HexEncryptionKey, false, messageID)
	tssService, err := tss.NewService(messenger, localStateAccessor, createPreParam)
	if err != nil {
		return nil, fmt.Errorf("create TSS service: %w", err)
	}
	return tssService, nil
}

func (s *WorkerService) startMessageDownload(serverURL, session, key, hexEncryptionKey string, tssService tss.Service, messageID string) (chan struct{}, *sync.WaitGroup) {
	s.logger.WithFields(logrus.Fields{
		"session": session,
		"key":     key,
	}).Info("Start downloading messages")

	endCh := make(chan struct{})
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go s.downloadMessages(serverURL, session, key, hexEncryptionKey, tssService, endCh, messageID, wg)
	return endCh, wg
}

func (s *WorkerService) downloadMessages(server, session, localPartyID, hexEncryptionKey string, tssServerImp tss.Service, endCh chan struct{}, messageID string, wg *sync.WaitGroup) {
	var messageCache sync.Map
	defer wg.Done()
	logger := s.logger.WithFields(logrus.Fields{
		"session":        session,
		"local_party_id": localPartyID,
	})
	logger.Info("Start downloading messages from : ", server)
	relayClient := relay.NewRelayClient(server)
	for {
		select {
		case <-endCh: // we are done
			logger.Info("Stop downloading messages")
			return
		case <-time.After(time.Second):
			messages, err := relayClient.DownloadMessages(session, localPartyID, messageID)
			if err != nil {
				logger.Errorf("Failed to get messages: %v", err)
				continue
			}
			for _, message := range messages {
				cacheKey := fmt.Sprintf("%s-%s-%s", session, localPartyID, message.Hash)
				if messageID != "" {
					cacheKey = fmt.Sprintf("%s-%s-%s-%s", session, localPartyID, messageID, message.Hash)
				}
				if _, found := messageCache.Load(cacheKey); found {
					logger.Infof("Message already applied, skipping,hash: %s", message.Hash)
					continue
				}

				decodedBody, err := base64.StdEncoding.DecodeString(message.Body)
				if err != nil {
					logger.Errorf("Failed to decode data: %v", err)
					continue
				}

				decryptedBody, err := decrypt(string(decodedBody), hexEncryptionKey)
				if err != nil {
					logger.Errorf("Failed to decrypt data: %v", err)
					continue
				}

				if err := tssServerImp.ApplyData(decryptedBody); err != nil {
					logger.Errorf("Failed to apply data: %v", err)
					continue
				}

				messageCache.Store(cacheKey, true)
				if err := relayClient.DeleteMessageFromServer(session, localPartyID, message.Hash, messageID); err != nil {
					logger.Errorf("Failed to delete message: %v", err)
				}
			}
		}
	}
}

func decrypt(cipherText, hexKey string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	result = ""
	err = nil
	var block cipher.Block
	key, decodeErr := hex.DecodeString(hexKey)
	if decodeErr != nil {
		err = decodeErr
		return
	}
	cipherByte := []byte(cipherText)
	if block, err = aes.NewCipher(key); err != nil {
		return
	}

	if len(cipherByte) < aes.BlockSize {
		err = fmt.Errorf("ciphertext too short")
		return
	}

	iv := cipherByte[:aes.BlockSize]
	cipherByte = cipherByte[aes.BlockSize:]
	cbc := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(cipherByte))
	cbc.CryptBlocks(plaintext, cipherByte)
	plaintext, err = unpad(plaintext)
	if err != nil {
		return
	}
	result = string(plaintext)
	return
}

func unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("unpad: input data is empty")
	}

	paddingLen := int(data[length-1])
	if paddingLen > length || paddingLen == 0 {
		return nil, errors.New("unpad: invalid padding length")
	}

	for i := 0; i < paddingLen; i++ {
		if data[length-1-i] != byte(paddingLen) {
			return nil, errors.New("unpad: invalid padding")
		}
	}

	return data[:length-paddingLen], nil
}

func (s *WorkerService) JoinKeySign(req types.KeysignRequest) (map[string]tss.KeysignResponse, error) {

	s.logger.WithFields(logrus.Fields{
		"derivePath": req.DerivePath,
		"isECDSA":    req.IsECDSA,
		"messages":   req.Messages,
		"publicKey":  req.PublicKey,
		"session":    req.SessionID,
	}).Debug("JoinKeySign params received")

	result := map[string]tss.KeysignResponse{}
	keyFolder := s.cfg.VaultsFilePath
	serverURL := s.cfg.Relay.Server
	localStateAccessor, err := relay.NewLocalStateAccessorImp(keyFolder, req.PublicKey, req.VaultPassword, s.blockStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to create localStateAccessor: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"vault_public_key":   localStateAccessor.Vault.PublicKeyEcdsa, // Should match policy
		"request_public_key": req.PublicKey,
		"local_party_id":     localStateAccessor.Vault.LocalPartyId,
	}).Info("Loaded vault for signing")

	keyShare, err := localStateAccessor.GetLocalState(req.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get key share: %w", err)
	}
	s.logger.WithField("key_share_length", len(keyShare)).Info("Loaded key share")

	localPartyId := localStateAccessor.Vault.LocalPartyId
	server := relay.NewRelayClient(serverURL)

	// Let's register session here
	if req.StartSession {
		if err := server.StartSession(req.SessionID, req.Parties); err != nil {
			return nil, fmt.Errorf("failed to start session: %w", err)
		}
	} else {
		if err := server.RegisterSessionWithRetry(req.SessionID, localPartyId); err != nil {
			return nil, fmt.Errorf("failed to register session: %w", err)
		}
	}
	// wait longer for keysign start
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute+3*time.Second)
	defer cancel()

	partiesJoined, err := server.WaitForSessionStart(ctx, req.SessionID)
	s.logger.WithFields(logrus.Fields{
		"session":        req.SessionID,
		"parties_joined": partiesJoined,
	}).Info("Session started")

	if err != nil {
		return nil, fmt.Errorf("failed to wait for session start: %w", err)
	}

	for _, message := range req.Messages {
		var signature *tss.KeysignResponse
		for attempt := 0; attempt < 3; attempt++ {
			signature, err = s.keysignWithRetry(serverURL,
				localPartyId,
				req,
				partiesJoined,
				message,
				localStateAccessor.Vault.PublicKeyEddsa,
				localStateAccessor)
			if err == nil {
				break
			}
		}
		if err != nil {
			return result, err
		}
		if signature == nil {
			return result, fmt.Errorf("signature is nil")
		}
		result[message] = *signature
	}

	if err := server.CompleteSession(req.SessionID, localPartyId); err != nil {
		s.logger.WithFields(logrus.Fields{
			"session": req.SessionID,
			"error":   err,
		}).Error("Failed to complete session")
	}

	return result, nil
}

// TODO: define types.TxQueueKeysignRequest
func (s *WorkerService) JoinKeySignInTxQueue(req types.KeysignRequest) (map[string]tss.KeysignResponse, error) {
	s.logger.WithFields(logrus.Fields{
		"session id":  req.SessionID,
		"public key":  req.PublicKey,
		"is ECDSA":    req.IsECDSA,
		"derive path": req.DerivePath,
	}).Debug("Join keysign in tx queue")

	localStateAccessor, err := relay.NewLocalStateAccessorImp(s.cfg.VaultsFilePath, req.PublicKey, req.VaultPassword, s.blockStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to init localStateAccessor: %w", err)
	}
	localPartyId := localStateAccessor.Vault.LocalPartyId

	s.logger.WithFields(logrus.Fields{
		"vault public key ECDSA": localStateAccessor.Vault.PublicKeyEcdsa,
		"vault public key EDDSA": localStateAccessor.Vault.PublicKeyEddsa,
		"vault local party ID":   localPartyId,
	}).Info("Loaded vault for signing")

	keyShare, err := localStateAccessor.GetLocalState(req.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get key share: %w", err)
	}
	s.logger.WithField("key share length", len(keyShare)).Info("Loaded key share")

	txQueueClient := relay.NewTxQueueClient(s.cfg.TxQueue.Server)

	if req.StartSession {
		// accept tx proposal (session start)
		acceptTxRequest := relay.AcceptTxRequest{
			SessionID: req.SessionID,
			SignerID:  req.Parties[1], // TODO: get the new party id
		}
		if err := txQueueClient.AcceptTxProposalWithSessionStart(acceptTxRequest); err != nil {
			return nil, fmt.Errorf("failed to accept tx proposal: %w", err)
		}
	} else {
		// propose tx (register session)
		proposeTxRequest := relay.ProposeTxRequest{
			SessionID:      req.SessionID,
			PublicKey:      req.PublicKey,
			IsECDSA:        req.IsECDSA,
			LeaderSignerID: localPartyId,
			Threshold:      2,
			Chain:          "Ethereum",
			// TODO: just for testing
			TxPayload: relay.TxPayload{
				From:  "0xe5F238C95142be312852e864B830daADB9B7D290",
				To:    "0xfA0635a1d083D0bF377EFbD48DA46BB17e0106cA",
				Data:  "0x00",
				Value: "10000000",
			},
		}
		if err := txQueueClient.ProposeTxWithSessionRegistration(proposeTxRequest); err != nil {
			return nil, fmt.Errorf("failed to propose tx: %w", err)
		}
	}

	// wait for tx proposal acceptance and tx hash (keysign to start)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute+3*time.Second)
	defer cancel()
	txProposal, err := txQueueClient.WaitForAcceptedTxProposalWithHash(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for tx proposal start: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"session ID": req.SessionID,
		"signers":    txProposal.Signers,
		"tx hash":    txProposal.TxHash,
		"nonce":      txProposal.TxPayload.Nonce,
		"gas price":  txProposal.TxPayload.GasPrice,
		"gas limit":  txProposal.TxPayload.Gas,
	}).Info("Tx proposal accepted")

	result := map[string]tss.KeysignResponse{}

	req.Messages = []string{txProposal.TxHash}

	for _, message := range req.Messages {
		var signature *tss.KeysignResponse
		for attempt := 0; attempt < 3; attempt++ {
			signature, err = s.keysignWithRetry(s.cfg.TxQueue.Server, localPartyId, req, txProposal.Signers, message, localStateAccessor.Vault.PublicKeyEddsa, localStateAccessor)
			if err == nil {
				break
			}
		}
		if err != nil {
			return result, err
		}
		if signature == nil {
			return result, fmt.Errorf("signature is nil")
		}
		result[message] = *signature
	}

	if err := txQueueClient.CompleteSession(req.SessionID, localPartyId); err != nil {
		s.logger.WithFields(logrus.Fields{
			"session": req.SessionID,
			"error":   err,
		}).Error("Failed to complete session")
	}

	return result, nil
}

func (s *WorkerService) keysignWithRetry(
	serverURL,
	localPartyId string,
	req types.KeysignRequest,
	partiesJoined []string,
	msg string,
	publicKeyEdDSA string,
	localStateAccessor *relay.LocalStateAccessorImp,
) (*tss.KeysignResponse, error) {
	md5Hash := md5.Sum([]byte(msg))
	messageID := hex.EncodeToString(md5Hash[:])
	s.logger.Infoln("Start keysign for message: ", messageID)
	tssService, err := s.createTSSService(serverURL, req.SessionID, req.HexEncryptionKey, localStateAccessor, false, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to create TSS service: %w", err)
	}
	msgBuf, err := hex.DecodeString(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}
	messageToSign := base64.StdEncoding.EncodeToString(msgBuf)
	endCh, wg := s.startMessageDownload(serverURL, req.SessionID, localPartyId, req.HexEncryptionKey, tssService, messageID)

	var signature *tss.KeysignResponse
	if req.IsECDSA {
		signature, err = tssService.KeysignECDSA(&tss.KeysignRequest{
			PubKey:               req.PublicKey,
			MessageToSign:        messageToSign,
			LocalPartyKey:        localPartyId,
			KeysignCommitteeKeys: strings.Join(partiesJoined, ","),
			DerivePath:           req.DerivePath,
		})
	} else {
		signature, err = tssService.KeysignEdDSA(&tss.KeysignRequest{
			PubKey:               publicKeyEdDSA, // request public key should be EdDSA public key
			MessageToSign:        messageToSign,
			LocalPartyKey:        localPartyId,
			KeysignCommitteeKeys: strings.Join(partiesJoined, ","),
			DerivePath:           req.DerivePath,
		})
	}

	txQueueClient := relay.NewTxQueueClient(serverURL)
	if err == nil {
		if err := txQueueClient.MarkKeysignComplete(req.SessionID, messageID, *signature); err != nil {
			s.logger.Errorf("fail to mark keysign complete: %v", err)
		}
	} else {
		s.logger.Errorf("fail to key sign: %v", err)
		sigResp, err := txQueueClient.CheckKeysignComplete(req.SessionID, messageID)
		if err == nil && sigResp != nil {
			signature = sigResp
		}
	}
	close(endCh)
	wg.Wait()
	return signature, nil
}
