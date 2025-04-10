package syncer

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/vultisig/vultisigner/internal/types"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestCreatePolicySync(t *testing.T) {
	testCases := []struct {
		name           string
		policy         types.PluginPolicy
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name: "Successful creation sync",
			policy: types.PluginPolicy{
				ID:            "policy-1",
				PublicKey:     "test-public-key",
				IsEcdsa:       true,
				ChainCodeHex:  "test-chain-code-hex",
				DerivePath:    "test-derive-path",
				PluginID:      "test-plugin-id",
				PluginVersion: "test-plugin-version",
				PolicyVersion: "test-policy-version",
				PluginType:    "test-plugin",
				Signature:     "test-signature",
				Policy:        json.RawMessage(`{"id":"policy-1"}`),
				Active:        true,
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/plugin/policy", r.URL.Path)

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var sentPolicy types.PluginPolicy
				require.NoError(t, json.Unmarshal(body, &sentPolicy))
				require.Equal(t, "policy-1", sentPolicy.ID)
				require.Equal(t, "test-plugin", sentPolicy.PluginType)

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success"}`))
			},
			expectedErrMsg: "",
			wantErr:        false,
		},
		{
			name:   "Error creating policy",
			policy: types.PluginPolicy{},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"server error"}`))
			},
			wantErr:        true,
			expectedErrMsg: "fail to sync policy with verifier server",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			server := httptest.NewServer(http.HandlerFunc(tc.serverResponse))
			defer server.Close()

			serverURL := server.URL
			hostPort := strings.TrimPrefix(serverURL, "http://")
			parts := strings.Split(hostPort, ":")
			host := parts[0]
			port, err := strconv.ParseInt(parts[1], 10, 64)
			require.NoError(t, err)

			syncer := Syncer{
				logger:     logrus.StandardLogger(),
				serverAddr: fmt.Sprintf("http://%s:%d", host, port),
				client:     server.Client(),
			}

			err = syncer.CreatePolicySync(tc.policy)

			if tc.wantErr {
				require.Error(t, err)
				if tc.expectedErrMsg != "" {
					require.Contains(t, err.Error(), tc.expectedErrMsg)
				}
			} else {
				require.NoError(t, err)
			}

		})
	}
}

func TestUpdatePolicySync(t *testing.T) {
	testCases := []struct {
		name           string
		policy         types.PluginPolicy
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name: "Successful update sync",
			policy: types.PluginPolicy{
				ID:            "policy-1",
				PublicKey:     "test-public-key",
				IsEcdsa:       true,
				ChainCodeHex:  "test-chain-code-hex",
				DerivePath:    "test-derive-path",
				PluginID:      "test-plugin-id",
				PluginVersion: "test-plugin-version",
				PolicyVersion: "test-policy-version",
				PluginType:    "test-plugin",
				Signature:     "test-signature",
				Policy:        json.RawMessage(`{"id":"policy-1"}`),
				Active:        true,
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(t, "/plugin/policy", r.URL.Path)

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var sentPolicy types.PluginPolicy
				require.NoError(t, json.Unmarshal(body, &sentPolicy))
				require.Equal(t, "policy-1", sentPolicy.ID)
				require.Equal(t, "test-plugin", sentPolicy.PluginType)

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success"}`))
			},
			expectedErrMsg: "",
			wantErr:        false,
		},

		{
			name:   "Error updating policy",
			policy: types.PluginPolicy{},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":"server error"}`))
			},
			wantErr:        true,
			expectedErrMsg: "fail to sync policy with verifier server",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			server := httptest.NewServer(http.HandlerFunc(tc.serverResponse))
			defer server.Close()

			serverURL := server.URL
			hostPort := strings.TrimPrefix(serverURL, "http://")
			parts := strings.Split(hostPort, ":")
			host := parts[0]
			port, err := strconv.ParseInt(parts[1], 10, 64)
			require.NoError(t, err)

			syncer := Syncer{
				logger:     logrus.StandardLogger(),
				serverAddr: fmt.Sprintf("http://%s:%d", host, port),
				client:     server.Client(),
			}

			err = syncer.UpdatePolicySync(tc.policy)

			if tc.wantErr {
				require.Error(t, err)
				if tc.expectedErrMsg != "" {
					require.Contains(t, err.Error(), tc.expectedErrMsg)
				}
			} else {
				require.NoError(t, err)
			}

		})
	}

}

func TestDeletePolicySync(t *testing.T) {
	testCases := []struct {
		name           string
		policyID       string
		signature      string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name:      "Successful delete sync",
			policyID:  "policy-1",
			signature: "test-signature",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, "/plugin/policy/policy-1", r.URL.Path)

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var reqBody DeleteRequestBody
				require.NoError(t, json.Unmarshal(body, &reqBody))

				require.Equal(t, "test-signature", reqBody.Signature)

				w.WriteHeader(http.StatusNoContent)
			},
			expectedErrMsg: "",
			wantErr:        false,
		},
		{
			name:      "Error deleting policy",
			policyID:  "policy-2",
			signature: "signature-2",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"server error"}`))
			},
			wantErr:        true,
			expectedErrMsg: "fail to delete policy on verifier server",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			server := httptest.NewServer(http.HandlerFunc(tc.serverResponse))
			defer server.Close()

			serverURL := server.URL
			hostPort := strings.TrimPrefix(serverURL, "http://")
			parts := strings.Split(hostPort, ":")
			host := parts[0]
			port, err := strconv.ParseInt(parts[1], 10, 64)
			require.NoError(t, err)

			syncer := Syncer{
				logger:     logrus.StandardLogger(),
				serverAddr: fmt.Sprintf("http://%s:%d", host, port),
				client:     server.Client(),
			}

			err = syncer.DeletePolicySync(tc.policyID, tc.signature)

			if tc.wantErr {
				require.Error(t, err)
				if tc.expectedErrMsg != "" {
					require.Contains(t, err.Error(), tc.expectedErrMsg)
				}
			} else {
				require.NoError(t, err)
			}

		})
	}
}

func TestSyncTransaction(t *testing.T) {
	testCases := []struct {
		name           string
		action         Action
		jwtToken       string
		tx             types.TransactionHistory
		serverResponse func(w http.ResponseWriter, r *http.Request)
		wantErr        bool
		expectedErrMsg string
	}{
		{
			name:     "Successful create transaction sync",
			action:   CreateAction,
			jwtToken: "test-jwt-token-1",
			tx: types.TransactionHistory{
				TxHash: "txhash-1",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/sync/transaction", r.URL.Path)
				require.Equal(t, "Bearer test-jwt-token-1", r.Header.Get("Authorization"))

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var sentTx types.TransactionHistory
				err = json.Unmarshal(body, &sentTx)
				require.NoError(t, err)

				require.Equal(t, "txhash-1", sentTx.TxHash)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success"}`))
			},
			expectedErrMsg: "",
			wantErr:        false,
		},
		{
			name:     "Successful update transaction sync",
			action:   UpdateAction,
			jwtToken: "test-jwt-token-1",
			tx: types.TransactionHistory{
				TxHash: "txhash-1",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(t, "/sync/transaction", r.URL.Path)
				require.Equal(t, "Bearer test-jwt-token-1", r.Header.Get("Authorization"))

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var sentTx types.TransactionHistory
				err = json.Unmarshal(body, &sentTx)
				require.NoError(t, err)

				require.Equal(t, "txhash-1", sentTx.TxHash)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success"}`))
			},
			expectedErrMsg: "",
			wantErr:        false,
		},
		{
			name:     "Server error transaction sync",
			action:   CreateAction,
			jwtToken: "test-jwt-token-1",
			tx: types.TransactionHistory{
				TxHash: "txhash-1",
			},
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"server error"}`))
			},
			wantErr:        true,
			expectedErrMsg: "fail to sync transaction with verifier server",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			server := httptest.NewServer(http.HandlerFunc(tc.serverResponse))
			defer server.Close()

			serverURL := server.URL
			hostPort := strings.TrimPrefix(serverURL, "http://")
			parts := strings.Split(hostPort, ":")
			host := parts[0]
			port, err := strconv.ParseInt(parts[1], 10, 64)
			require.NoError(t, err)

			syncer := Syncer{
				logger:     logrus.StandardLogger(),
				serverAddr: fmt.Sprintf("http://%s:%d", host, port),
				client:     server.Client(),
			}

			err = syncer.SyncTransaction(tc.action, tc.jwtToken, tc.tx)

			if tc.wantErr {
				require.Error(t, err)
				if tc.expectedErrMsg != "" {
					require.Contains(t, err.Error(), tc.expectedErrMsg)
				}
			} else {
				require.NoError(t, err)
			}

		})
	}
}

func TestRetryMechanism(t *testing.T) {
	retryCount := 0
	maxRetries := 3

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retryCount++
		if retryCount < maxRetries {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"server error"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	serverURL := server.URL
	hostPort := strings.TrimPrefix(serverURL, "http://")
	parts := strings.Split(hostPort, ":")
	host := parts[0]
	port, err := strconv.ParseInt(parts[1], 10, 64)
	require.NoError(t, err)

	syncer := Syncer{
		logger:     logrus.StandardLogger(),
		serverAddr: fmt.Sprintf("http://%s:%d", host, port),
		client:     server.Client(),
	}

	policy := types.PluginPolicy{
		ID:         "retry-test",
		PluginType: "test-plugin",
	}

	start := time.Now()
	err = syncer.CreatePolicySync(policy)
	duration := time.Since(start)

	require.NoError(t, err)
	minimumExpectedDuration := 3 * initialBackoff
	require.GreaterOrEqual(t, duration, minimumExpectedDuration, "Test completed too quickly, expected multiple retries")

	require.Equal(t, maxRetries, retryCount, "Expected exactly %d attempts", maxRetries)
}
