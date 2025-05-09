package syncer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultiserver-plugin/internal/types"
)

const (
	defaultTimeout      = 10 * time.Second
	policyEndpoint      = "/plugin/policy"
	transactionEndpoint = "/sync/transaction"

	// Retry configuration
	maxRetries     = 3
	initialBackoff = 100 * time.Millisecond
)

type Action int
const (
	CreateAction Action = iota
	UpdateAction
)

type PolicySyncer interface {
	CreatePolicySync(policy types.PluginPolicy) error
	UpdatePolicySync(policy types.PluginPolicy) error
	DeletePolicySync(policyID, signature string) error
	SyncTransaction(action Action, jwtToken string, tx types.TransactionHistory) error
}

type Syncer struct {
	logger     *logrus.Logger
	client     *http.Client
	serverAddr string
}

func NewPolicySyncer(logger *logrus.Logger, serverHost string, serverPort int64) PolicySyncer {
	return &Syncer{
		logger: logger,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
		serverAddr: fmt.Sprintf("http://%s:%d", serverHost, serverPort),
	}
}

func (s *Syncer) CreatePolicySync(policy types.PluginPolicy) error {
	s.logger.WithFields(logrus.Fields{
		"policy_id":   policy.ID,
		"plugin_type": policy.PluginType,
	}).Info("Starting policy creation sync")

	return s.retryWithBackoff("CreatePolicySync", func() error {
		policyBytes, err := json.Marshal(policy)
		if err != nil {
			return fmt.Errorf("fail to marshal policy: %w", err)
		}

		url := s.serverAddr + policyEndpoint

		resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(policyBytes))
		if err != nil {
			return fmt.Errorf("fail to sync policy with verifier server: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			s.logger.WithFields(logrus.Fields{
				"status_code": resp.StatusCode,
				"body":        string(body),
				"policy_id":   policy.ID,
			}).Error("Failed to sync create policy")
			return fmt.Errorf("fail to sync policy with verifier server: status: %d", resp.StatusCode)
		}

		s.logger.WithFields(logrus.Fields{
			"policy_id": policy.ID,
		}).Info("Successfully sync created policy")

		return nil
	})
}

func (s *Syncer) UpdatePolicySync(policy types.PluginPolicy) error {
	s.logger.WithFields(logrus.Fields{
		"policy_id":   policy.ID,
		"plugin_type": policy.PluginType,
	}).Info("Starting policy update sync")

	return s.retryWithBackoff("UpdatePolicySync", func() error {
		policyBytes, err := json.Marshal(policy)
		if err != nil {
			return fmt.Errorf("fail to marshal policy: %w", err)
		}

		url := s.serverAddr + policyEndpoint

		req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(policyBytes))
		if err != nil {
			return fmt.Errorf("fail to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			return fmt.Errorf("fail to sync policy with verifier server: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			s.logger.WithFields(logrus.Fields{
				"status_code": resp.StatusCode,
				"body":        string(body),
				"policy_id":   policy.ID,
			}).Error("Failed to sync update policy")
			return fmt.Errorf("fail to sync policy with verifier server, status: %d", resp.StatusCode)
		}

		s.logger.WithFields(logrus.Fields{
			"policy_id": policy.ID,
		}).Info("Successfully sync updated policy")

		return nil
	})
}

type DeleteRequestBody struct {
	Signature string `json:"signature"`
}

func (s *Syncer) DeletePolicySync(policyID, signature string) error {
	s.logger.WithFields(logrus.Fields{
		"policy_id": policyID,
	}).Info("Starting policy delete sync")

	return s.retryWithBackoff("DeletePolicySync", func() error {
		reqBody := DeleteRequestBody{
			Signature: signature,
		}
		reqBodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("fail to marshal request body: %w", err)
		}

		url := s.serverAddr + policyEndpoint + "/" + policyID

		req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(reqBodyBytes))
		if err != nil {
			return fmt.Errorf("fail to create request, err: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			return fmt.Errorf("fail to delete policy on verifier server, err: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusNoContent {
			s.logger.WithFields(logrus.Fields{
				"status_code": resp.StatusCode,
				"body":        string(body),
				"policy_id":   policyID,
			}).Error("Failed to sync delete policy")
			return fmt.Errorf("fail to delete policy on verifier server, status: %d", resp.StatusCode)
		}

		s.logger.WithFields(logrus.Fields{
			"policy_id": policyID,
		}).Info("Successfully sync deleted policy")

		return nil
	})
}

func (s *Syncer) SyncTransaction(action Action, jwtToken string, tx types.TransactionHistory) error {
	s.logger.WithFields(logrus.Fields{
		"tx_id":   tx.ID,
		"tx_hash": tx.TxHash,
	}).Info("Starting tx sync")

	return s.retryWithBackoff("SyncTransaction", func() error {
		txBytes, err := json.Marshal(tx)
		if err != nil {
			return fmt.Errorf("fail to marshal transaction: %w", err)
		}
		url := s.serverAddr + transactionEndpoint
		var method string
		switch action {
		case CreateAction:
			method = http.MethodPost
		case UpdateAction:
			method = http.MethodPut
		}

		req, err := http.NewRequest(method, url, bytes.NewBuffer(txBytes))
		if err != nil {
			return fmt.Errorf("fail to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwtToken))

		resp, err := s.client.Do(req)
		if err != nil {
			return fmt.Errorf("fail to sync transaction on verifier server: %w", err)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			s.logger.WithFields(logrus.Fields{
				"status_code": resp.StatusCode,
				"body":        string(body),
				"tx_id":       tx.ID,
			}).Error("Failed to sync update policy")
			return fmt.Errorf("fail to sync transaction with verifier server, status: %d", resp.StatusCode)
		}

		s.logger.WithFields(logrus.Fields{
			"tx_id": tx.ID,
		}).Info("Successfully sync transaction")

		return nil
	})
}

// retryWithBackoff attempts to execute the given operation with exponential backoff
func (s *Syncer) retryWithBackoff(operation string, fn func() error) error {
	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			s.logger.WithFields(logrus.Fields{
				"attempt":   attempt,
				"backoff":   backoff.String(),
				"operation": operation,
			}).Debug("Retrying sync")

			time.Sleep(backoff)
			backoff *= 2
		}

		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		s.logger.WithFields(logrus.Fields{
			"attempt":   attempt,
			"error":     err.Error(),
			"operation": operation,
		}).Warn("Sync failed, will retry")
	}

	return fmt.Errorf("sync failed: %w", lastErr)
}
