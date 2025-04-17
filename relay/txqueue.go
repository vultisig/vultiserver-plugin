package relay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vultisig/mobile-tss-lib/tss"
)

type ProposeTxRequest struct {
	SessionID      string    `json:"session_id" validate:"required"`
	PublicKey      string    `json:"public_key" validate:"required"`
	IsECDSA        bool      `json:"is_ecdsa" validate:"required"`
	LeaderSignerID string    `json:"leader_signer_id" validate:"required"`
	Threshold      uint8     `json:"threshold" validate:"required,gte=0"`
	Chain          string    `json:"chain" validate:"required"`
	TxPayload      TxPayload `json:"tx_payload" validate:"required"`
}

type AcceptTxRequest struct {
	SessionID string `json:"session_id" validate:"required"`
	SignerID  string `json:"signer_id" validate:"required"`
}

type getTxProposalRequest struct {
	ID             uint64    `json:"id" validate:"required"`
	Status         uint8     `json:"status" validate:"required"`
	ErrorMessage   string    `json:"error_message" validate:"required"`
	SessionID      string    `json:"session_id" validate:"required"`
	PublicKey      string    `json:"public_key" validate:"required"`
	IsECDSA        bool      `json:"is_ecdsa" validate:"required"`
	Threshold      uint8     `json:"threshold" validate:"required,gte=0"`
	LeaderSignerID string    `json:"leader_signer_id" validate:"required"`
	Signers        []string  `json:"signers" validate:"required"`
	Chain          string    `json:"chain" validate:"required"`
	TxPayload      TxPayload `json:"tx_payload" validate:"required"`
	TxHash         string    `json:"tx_hash" validate:"required,hexadecimal"`
	Signature      string    `json:"signature" validate:"required,hexadecimal"`
	CreatedAt      time.Time `json:"created_at" validate:"required"`
	UpdatedAt      time.Time `json:"updated_at" validate:"required"`
}

// TODO: add multichain support

type TxPayload struct {
	From  string `json:"from" validate:"required,hexadecimal"`
	To    string `json:"to,omitempty" validate:"omitempty,hexadecimal"`
	Data  string `json:"data" validate:"required,hexadecimal"`
	Value string `json:"value" validate:"required,hexadecimal"`
	// dynamic fields - optional during creation
	Nonce    uint64 `json:"nonce" validate:"omitempty"`
	Gas      uint64 `json:"gas" validate:"omitempty"`
	GasPrice string `json:"gas_price" validate:"omitempty,hexadecimal"`
}

type TxQueueClient struct {
	url        string
	httpClient http.Client
	logger     *logrus.Logger
}

func NewTxQueueClient(url string) *TxQueueClient {
	return &TxQueueClient{
		url:        url,
		httpClient: http.Client{Timeout: 5 * time.Second},
		logger:     logrus.WithField("service", "tx-queue-client").Logger,
	}
}

func (c *TxQueueClient) bodyCloser(body io.ReadCloser) {
	if body != nil {
		if err := body.Close(); err != nil {
			c.logger.Error("Failed to close body,err:", err)
		}
	}
}

func (c *TxQueueClient) ProposeTxWithSessionRegistration(payload ProposeTxRequest) error {
	url := c.url + "/v1/proposals"

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("fail to marshal payload: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		c.logger.Error("fail to propose tx and register session:", "error", err)
		return fmt.Errorf("fail to register session: %w", err)
	}
	defer c.bodyCloser(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		c.logger.Error("fail to register session:", "status", resp.Status)
		return fmt.Errorf("fail to register session: %s", resp.Status)
	}
	return nil
}

func (c *TxQueueClient) AcceptTxProposalWithSessionStart(payload AcceptTxRequest) error {
	url := c.url + "/v1/proposals/accept"

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("fail to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("fail to accept tx proposal and start session: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("fail to start session ", "error", err)
		return fmt.Errorf("fail to start session: %w", err)
	}
	defer c.bodyCloser(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to start session: %s", resp.Status)
	}
	return nil
}

func (c *TxQueueClient) WaitForAcceptedTxProposalWithHash(ctx context.Context, sessionID string) (*getTxProposalRequest, error) {
	url := c.url + "/v1/proposals/" + sessionID

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			resp, err := c.httpClient.Get(url)
			if err != nil {
				return nil, fmt.Errorf("fail to get session: %w", err)
			}
			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("fail to get session: %s", resp.Status)
			}

			var txProposal getTxProposalRequest
			buff, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("fail to read session body: %w", err)
			}
			c.bodyCloser(resp.Body)
			if err := json.Unmarshal(buff, &txProposal); err != nil {
				return nil, fmt.Errorf("fail to unmarshal session body: %w", err)
			}

			// check if tx proposal is accepted and is ready to be signed
			if len(txProposal.Signers) >= int(txProposal.Threshold) && txProposal.TxHash != "" {
				c.logger.WithFields(logrus.Fields{
					"session": sessionID,
					"parties": txProposal.Signers,
				}).Info("All parties joined")
				return &txProposal, nil
			}

			c.logger.WithFields(logrus.Fields{
				"session": sessionID,
			}).Info("Waiting for someone to start session")

			// backoff
			time.Sleep(1 * time.Second)
		}
	}
}

func (c *TxQueueClient) CompleteSession(sessionID, localPartyID string) error {
	url := c.url + "/complete/" + sessionID

	parties := []string{localPartyID}
	body, err := json.Marshal(parties)
	if err != nil {
		return fmt.Errorf("fail to complete session: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("fail to complete session: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fail to complete session: %w", err)
	}
	defer c.bodyCloser(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to complete session: %s", resp.Status)
	}
	return nil
}

func (c *TxQueueClient) MarkKeysignComplete(sessionID string, messageID string, sig tss.KeysignResponse) error {
	url := c.url + "/complete/" + sessionID + "/keysign"

	body, err := json.Marshal(sig)
	if err != nil {
		return fmt.Errorf("fail to marshal keysign to json: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("fail to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("message_id", messageID)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fail to mark keysign complete: %w", err)
	}
	defer c.bodyCloser(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to mark keysign complete: %s", resp.Status)
	}
	return nil
}

func (c *TxQueueClient) CheckKeysignComplete(sessionID string, messageID string) (*tss.KeysignResponse, error) {
	url := c.url + "/complete/" + sessionID + "/keysign"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("fail to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("message_id", messageID)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fail to check keysign complete: %w", err)
	}
	defer c.bodyCloser(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fail to check keysign complete: %s", resp.Status)
	}
	var sig tss.KeysignResponse
	if err := json.NewDecoder(resp.Body).Decode(&sig); err != nil {
		return nil, fmt.Errorf("fail to unmarshal keysign response: %w", err)
	}
	return &sig, nil
}
