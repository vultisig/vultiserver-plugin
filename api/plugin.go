package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vultisig/vultiserver-plugin/common"
	"github.com/vultisig/vultiserver-plugin/internal/jwt"
	"github.com/vultisig/vultiserver-plugin/internal/password"
	"github.com/vultisig/vultiserver-plugin/internal/sigutil"
	"github.com/vultisig/vultiserver-plugin/internal/tasks"
	"github.com/vultisig/vultiserver-plugin/internal/types"
	"github.com/vultisig/vultiserver-plugin/plugin"
	"github.com/vultisig/vultiserver-plugin/plugin/dca"
	"github.com/vultisig/vultiserver-plugin/plugin/payroll"

	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
)

func (s *Server) SignPluginMessages(c echo.Context) error {
	s.logger.Debug("PLUGIN SERVER: SIGN MESSAGES")

	var req types.PluginKeysignRequest
	if err := c.Bind(&req); err != nil {
		wrappedErr := fmt.Errorf("fail to parse request, err: %w", err)
		s.logger.Error(wrappedErr)
		message := map[string]interface{}{
			"error": wrappedErr.Error(),
		}
		return c.JSON(http.StatusBadRequest, message)
	}

	// Plugin-specific validations //TODO: maybe this is not okay since we do not want this kind of specific checks in verifier
	if len(req.Messages) != 1 {
		wrappedErr := fmt.Errorf("plugin signing requires exactly one message hash, current: %d", len(req.Messages))
		s.logger.Error(wrappedErr)
		message := map[string]interface{}{
			"error": wrappedErr.Error(),
		}
		return c.JSON(http.StatusBadRequest, message)
	}

	// Get policy from database
	policy, err := s.db.GetPluginPolicy(c.Request().Context(), req.PolicyID)
	if err != nil {
		wrappedErr := fmt.Errorf("failed to get policy from database: %w", err)
		s.logger.Error(wrappedErr)
		message := map[string]interface{}{
			"error": wrappedErr.Error(),
		}
		return c.JSON(http.StatusInternalServerError, message)
	}

	// Validate policy matches plugin
	if policy.PluginID != req.PluginID {
		mismatchErr := errors.New("policy plugin ID mismatch")
		s.logger.Error(mismatchErr)
		message := map[string]interface{}{
			"error": mismatchErr.Error(),
		}
		return c.JSON(http.StatusBadRequest, message)
	}

	// We re-init plugin as verification server doesn't have plugin defined
	var plg plugin.Plugin
	plg, err = s.initializePlugin(policy.PluginType)
	if err != nil {
		wrappedErr := fmt.Errorf("fail to initialize plugin: %w", err)
		s.logger.Error(wrappedErr)
		message := map[string]interface{}{
			"error": wrappedErr.Error(),
		}
		return c.JSON(http.StatusInternalServerError, message)
	}

	if err := plg.ValidateProposedTransactions(policy, []types.PluginKeysignRequest{req}); err != nil {
		wrappedErr := fmt.Errorf("fail to validate proposed transactions: %w", err)
		s.logger.Error(wrappedErr)
		message := map[string]interface{}{
			"error": wrappedErr.Error(),
		}
		return c.JSON(http.StatusBadRequest, message)
	}

	// Validate message hash matches transaction
	txHash, err := calculateTransactionHash(req.Transaction)
	if err != nil {
		wrappedErr := fmt.Errorf("fail to calculate transaction hash: %w", err)
		s.logger.Error(wrappedErr)
		message := map[string]interface{}{
			"error": wrappedErr.Error(),
		}
		return c.JSON(http.StatusInternalServerError, message)
	}
	if txHash != req.Messages[0] {
		wrappedErr := fmt.Errorf("message hash does not match transaction hash. expected %s, got %s", txHash, req.Messages[0])
		s.logger.Error(wrappedErr)
		message := map[string]interface{}{
			"error": wrappedErr.Error(),
		}
		return c.JSON(http.StatusInternalServerError, message)
	}

	// Reuse existing signing logic //TODO: why we return here with status OK
	result, err := s.redis.Get(c.Request().Context(), req.SessionID)
	if err == nil && result != "" {
		return c.NoContent(http.StatusOK)
	}

	if err := s.redis.Set(c.Request().Context(), req.SessionID, req.SessionID, 30*time.Minute); err != nil {
		//TODO: should we return error here or just log it
		s.logger.Errorf("fail to set session, err: %v", err)
	}

	filePathName := common.GetVaultBackupFilename(req.PublicKey)
	content, err := s.blockStorage.GetFile(filePathName)
	if err != nil {
		wrappedErr := fmt.Errorf("fail to read file: %s", filePathName)
		message := map[string]interface{}{
			"error": wrappedErr.Error(),
		}
		return c.JSON(http.StatusInternalServerError, message)
	}

	_, err = common.DecryptVaultFromBackup(req.VaultPassword, content)
	if err != nil {
		wrappedErr := fmt.Errorf("fail to decrypt file from the backup: %w", err)
		s.logger.Error(wrappedErr)
		message := map[string]interface{}{
			"error": wrappedErr.Error(),
		}
		return c.JSON(http.StatusInternalServerError, message)
	}

	req.Parties = []string{common.PluginPartyID, common.VerifierPartyID}

	buf, err := json.Marshal(req)
	if err != nil {
		wrappedErr := fmt.Errorf("fail to marshal request: %w", err)
		s.logger.Error(wrappedErr)
		message := map[string]interface{}{
			"error": wrappedErr.Error(),
		}
		return c.JSON(http.StatusInternalServerError, message)
	}

	// TODO: check if this is relevant
	// check that tx is done only once per period
	// should we also copy the db to the vultiserver, so that it can be used by the vultiserver (and use scheduler.go)? or query the blockchain?

	txToSign, err := s.db.GetTransactionByHash(c.Request().Context(), txHash)
	if err != nil {
		wrappedErr := fmt.Errorf("fail to get transaction by hash: %w", err)
		s.logger.Error(wrappedErr)
		message := map[string]interface{}{
			"error": wrappedErr.Error(),
		}
		return c.JSON(http.StatusInternalServerError, message)
	}

	s.logger.Debug("PLUGIN SERVER: KEYSIGN TASK")

	ti, err := s.client.EnqueueContext(c.Request().Context(),
		asynq.NewTask(tasks.TypeKeySign, buf),
		asynq.MaxRetry(0),
		asynq.Timeout(2*time.Minute),
		asynq.Retention(5*time.Minute),
		asynq.Queue(tasks.QUEUE_NAME))

	if err != nil {
		txToSign.Metadata["error"] = err.Error()
		if updateErr := s.db.UpdateTransactionStatus(c.Request().Context(), txToSign.ID, types.StatusSigningFailed, txToSign.Metadata); updateErr != nil {
			s.logger.Errorf("Failed to update transaction status: %v", updateErr)
		}
		wrappedErr := fmt.Errorf("fail to enqueue task: %w", err)
		s.logger.Error(wrappedErr)
		message := map[string]interface{}{
			"error": wrappedErr.Error(),
		}
		return c.JSON(http.StatusInternalServerError, message)
	}

	txToSign.Metadata["task_id"] = ti.ID
	if err := s.db.UpdateTransactionStatus(c.Request().Context(), txToSign.ID, types.StatusSigned, txToSign.Metadata); err != nil {
		s.logger.Errorf("Failed to update transaction with task ID: %v", err)
	}

	s.logger.Infof("Created transaction history for tx from plugin: %s...", req.Transaction[:min(20, len(req.Transaction))])

	return c.JSON(http.StatusOK, ti.ID)
}

func (s *Server) GetAllPluginPolicies(c echo.Context) error {
	publicKey := c.Request().Header.Get("public_key")
	if publicKey == "" {
		err := fmt.Errorf("missing required header: public_key")
		message := map[string]interface{}{
			"message": "failed to get policies",
			"error":   err.Error(),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, message)
	}

	pluginType := c.Request().Header.Get("plugin_type")
	if pluginType == "" {
		err := fmt.Errorf("missing required header: plugin_type")
		message := map[string]interface{}{
			"message": "failed to get policies",
			"error":   err.Error(),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, message)
	}

	policies, err := s.policyService.GetPluginPolicies(c.Request().Context(), publicKey, pluginType)
	if err != nil {
		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to get policies for public_key: %s", publicKey),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, policies)
}

func (s *Server) CreatePluginPolicy(c echo.Context) error {
	var policy types.PluginPolicy
	if err := c.Bind(&policy); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}

	// We re-init plugin as verification server doesn't have plugin defined
	var plg plugin.Plugin
	plg, err := s.initializePlugin(policy.PluginType)
	if err != nil {
		err = fmt.Errorf("failed to initialize plugin: %w", err)
		s.logger.Error(err)
		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to initialize plugin: %s", policy.PluginType),
		}
		return c.JSON(http.StatusBadRequest, message)
	}

	if err := plg.ValidatePluginPolicy(policy); err != nil {
		if errors.Unwrap(err) != nil {
			err = fmt.Errorf("failed to validate policy: %w", err)
			s.logger.Error(err)
			message := map[string]interface{}{
				"message": "failed to validate policy",
			}
			return c.JSON(http.StatusBadRequest, message)
		}

		err = fmt.Errorf("failed to validate policy: %w", err)
		s.logger.Error(err)
		message := map[string]interface{}{
			"error":   err.Error(), // only if error is not wrapped
			"message": "failed to validate policy",
		}

		return c.JSON(http.StatusBadRequest, message)
	}

	if policy.ID == "" {
		policy.ID = uuid.NewString()
	}

	if !s.verifyPolicySignature(policy, false) {
		s.logger.Error("invalid policy signature")
		message := map[string]interface{}{
			"message": "Authorization failed",
			"error":   "Invalid policy signature",
		}
		return c.JSON(http.StatusForbidden, message)
	}

	newPolicy, err := s.policyService.CreatePolicyWithSync(c.Request().Context(), policy)
	if err != nil {
		err = fmt.Errorf("failed to create plugin policy: %w", err)
		message := map[string]interface{}{
			"message": "failed to create policy",
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, newPolicy)
}

func (s *Server) UpdatePluginPolicyById(c echo.Context) error {
	var policy types.PluginPolicy
	if err := c.Bind(&policy); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}

	// We re-init plugin as verification server doesn't have plugin defined
	var plg plugin.Plugin
	plg, err := s.initializePlugin(policy.PluginType)
	if err != nil {
		if errors.Unwrap(err) != nil {
			err = fmt.Errorf("failed to initialize plugin: %w", err)

			message := map[string]interface{}{
				"message": fmt.Sprintf("failed to initialize plugin: %s", policy.PluginType),
			}

			s.logger.Error(err)
			return c.JSON(http.StatusBadRequest, message)
		}

		err = fmt.Errorf("failed to initialize plugin: %w", err)

		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to initialize plugin: %s", policy.PluginType),
			"error":   err.Error(),
		}

		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, message)
	}

	if err := plg.ValidatePluginPolicy(policy); err != nil {
		if errors.Unwrap(err) != nil {
			err = fmt.Errorf("failed to validate policy: %w", err)
			s.logger.Error(err)
			message := map[string]interface{}{
				"message": fmt.Sprintf("failed to validate policy: %s", policy.ID),
			}
			return c.JSON(http.StatusBadRequest, message)
		}

		err = fmt.Errorf("failed to validate policy: %w", err)
		s.logger.Error(err)
		message := map[string]interface{}{
			"error":   err.Error(), // only if error is not wrapped
			"message": fmt.Sprintf("failed to validate policy: %s", policy.ID),
		}
		return c.JSON(http.StatusBadRequest, message)
	}

	if !s.verifyPolicySignature(policy, true) {
		s.logger.Error("invalid policy signature")
		message := map[string]interface{}{
			"message": "Authorization failed",
			"error":   "Invalid policy signature",
		}
		return c.JSON(http.StatusForbidden, message)
	}

	updatedPolicy, err := s.policyService.UpdatePolicyWithSync(c.Request().Context(), policy)
	if err != nil {
		err = fmt.Errorf("failed to update plugin policy: %w", err)
		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to update policy: %s", policy.ID),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, updatedPolicy)
}

func (s *Server) DeletePluginPolicyById(c echo.Context) error {
	var reqBody struct {
		Signature string `json:"signature"`
	}

	if err := c.Bind(&reqBody); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}

	policyID := c.Param("policyId")
	if policyID == "" {
		err := fmt.Errorf("policy ID is required")
		message := map[string]interface{}{
			"message": "failed to delete policy",
			"error":   err.Error(),
		}
		s.logger.Error(err)

		return c.JSON(http.StatusBadRequest, message)
	}

	policy, err := s.policyService.GetPluginPolicy(c.Request().Context(), policyID)
	if err != nil {
		err = fmt.Errorf("failed to get policy: %w", err)
		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to get policy: %s", policyID),
			"error":   err.Error(),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	// This is because we have different signature stored in the database.
	policy.Signature = reqBody.Signature

	if !s.verifyPolicySignature(policy, true) {
		s.logger.Error("invalid policy signature")
		message := map[string]interface{}{
			"message": "Authorization failed",
			"error":   "Invalid policy signature",
		}
		return c.JSON(http.StatusForbidden, message)
	}

	if err := s.policyService.DeletePolicyWithSync(c.Request().Context(), policyID, reqBody.Signature); err != nil {
		err = fmt.Errorf("failed to delete policy: %w", err)
		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to delete policy: %s", policyID),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) GetPolicySchema(c echo.Context) error {
	pluginType := c.Request().Header.Get("plugin_type") // this is a unique identifier; this won't be needed once the DCA and Payroll are separate services
	if pluginType == "" {
		err := fmt.Errorf("missing required header: plugin_type")
		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to get policy schema for plugin: %s", pluginType),
			"error":   err.Error(),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, message)
	}

	keyPath := filepath.Join("plugin", pluginType, "dcaPluginUiSchema.json")

	jsonData, err := os.ReadFile(keyPath)
	if err != nil {
		message := map[string]interface{}{
			"message": fmt.Sprintf("missing schema for plugin: %s", pluginType),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, message)
	}

	var data map[string]interface{}
	jsonErr := json.Unmarshal([]byte(jsonData), &data)
	if jsonErr != nil {

		message := map[string]interface{}{
			"message": fmt.Sprintf("could not unmarshal json: %s", jsonErr),
			"error":   jsonErr.Error(),
		}
		s.logger.Error(jsonErr)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, data)
}

func (s *Server) GetPluginPolicyTransactionHistory(c echo.Context) error {
	policyID := c.Param("policyId")

	if policyID == "" {
		err := fmt.Errorf("policy ID is required")
		message := map[string]interface{}{
			"message": "failed to get policy",
			"error":   err.Error(),
		}
		return c.JSON(http.StatusBadRequest, message)
	}

	policyHistory, err := s.policyService.GetPluginPolicyTransactionHistory(c.Request().Context(), policyID)
	if err != nil {
		err = fmt.Errorf("failed to get policy history: %w", err)
		message := map[string]interface{}{
			"message": fmt.Sprintf("failed to get policy history: %s", policyID),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, policyHistory)
}

func (s *Server) initializePlugin(pluginType string) (plugin.Plugin, error) {
	switch pluginType {
	case payroll.PluginType:
		return payroll.NewPayrollPlugin(s.db, s.logger, s.pluginConfigs[payroll.PluginType])
	case dca.PluginType:
		return dca.NewDCAPlugin(s.db, s.logger, s.pluginConfigs[dca.PluginType])
	default:
		return nil, fmt.Errorf("unknown plugin type: %s", pluginType)
	}
}

func (s *Server) UserLogin(c echo.Context) error {
	var auth types.UserAuthDto
	if err := c.Bind(&auth); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "Invalid request"})
	}

	if err := c.Validate(&auth); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err.Error(),
		})
	}

	user, err := s.db.FindUserByName(c.Request().Context(), auth.Username)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Invalid credentials"})
	}
	if passwordValid := password.CheckPassword(auth.Password, user.Password); !passwordValid {
		return c.JSON(http.StatusUnauthorized, echo.Map{"message": "Invalid credentials"})
	}

	token, err := jwt.GenerateJWT(user.ID, s.cfg.UserAuth.JwtSecret)
	if err != nil {
		s.logger.Error("Failed to generate jwt", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

func (s *Server) GetLoggedUser(c echo.Context) error {
	user := c.Get("user")

	return c.JSON(http.StatusOK, user)
}

func (s *Server) GetPricing(c echo.Context) error {
	pricingID := c.Param("pricingId")
	if pricingID == "" {
		err := fmt.Errorf("pricing id is required")
		message := echo.Map{
			"message": "failed to get pricing",
			"error":   err.Error(),
		}
		s.logger.Error(err)

		return c.JSON(http.StatusBadRequest, message)
	}

	pricing, err := s.db.FindPricingById(c.Request().Context(), pricingID)
	if err != nil {
		message := echo.Map{
			"message": "failed to get pricing",
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, pricing)
}

func (s *Server) CreatePricing(c echo.Context) error {
	var pricing types.PricingCreateDto
	if err := c.Bind(&pricing); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}

	if err := c.Validate(&pricing); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err.Error(),
		})
	}

	created, err := s.db.CreatePricing(c.Request().Context(), pricing)
	if err != nil {
		message := echo.Map{
			"message": "failed to create pricing",
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, created)
}

func (s *Server) DeletePricing(c echo.Context) error {
	pricingID := c.Param("pricingId")
	if pricingID == "" {
		message := echo.Map{
			"message": "failed to delete pricing",
			"error":   "pricing id is required",
		}
		return c.JSON(http.StatusBadRequest, message)
	}

	err := s.db.DeletePricingById(c.Request().Context(), pricingID)
	if err != nil {
		message := echo.Map{
			"message": "failed to delete pricing",
			"error":   err.Error(),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) GetCategories(c echo.Context) error {
	categories, err := s.db.FindCategories(c.Request().Context())
	if err != nil {
		message := echo.Map{
			"message": "failed to get categories",
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, categories)
}

func (s *Server) GetTags(c echo.Context) error {
	tags, err := s.db.FindTags(c.Request().Context())
	if err != nil {
		message := echo.Map{
			"message": "failed to get tags",
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, tags)
}

func (s *Server) GetPlugins(c echo.Context) error {
	skip, err := strconv.Atoi(c.QueryParam("skip"))

	if err != nil {
		skip = 0
	}

	take, err := strconv.Atoi(c.QueryParam("take"))

	if err != nil {
		take = 999
	}

	sort := c.QueryParam("sort")

	filters := types.PluginFilters{
		Term:       common.GetQueryParam(c, "term"),
		TagID:      common.GetQueryParam(c, "tag_id"),
		CategoryID: common.GetQueryParam(c, "category_id"),
	}

	plugins, err := s.db.FindPlugins(c.Request().Context(), filters, skip, take, sort)

	if err != nil {
		message := echo.Map{
			"message": "failed to get plugins",
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, plugins)
}

func (s *Server) GetPlugin(c echo.Context) error {
	pluginID := c.Param("pluginId")
	if pluginID == "" {
		err := fmt.Errorf("plugin id is required")
		message := echo.Map{
			"message": "failed to get plugin",
			"error":   err.Error(),
		}
		s.logger.Error(err)

		return c.JSON(http.StatusBadRequest, message)
	}

	plugin, err := s.db.FindPluginById(c.Request().Context(), pluginID)
	if err != nil {
		message := echo.Map{
			"message": "failed to get plugin",
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, plugin)
}

func (s *Server) CreatePlugin(c echo.Context) error {
	var plugin types.PluginCreateDto
	if err := c.Bind(&plugin); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}

	if err := c.Validate(&plugin); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err.Error(),
		})
	}

	created, err := s.db.CreatePlugin(c.Request().Context(), plugin)
	if err != nil {
		message := echo.Map{
			"message": "failed to create plugin",
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, created)
}

func (s *Server) UpdatePlugin(c echo.Context) error {
	pluginID := c.Param("pluginId")
	if pluginID == "" {
		message := echo.Map{
			"message": "failed to delete plugin",
			"error":   "plugin id is required",
		}
		return c.JSON(http.StatusBadRequest, message)
	}

	var plugin types.PluginUpdateDto
	if err := c.Bind(&plugin); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}

	if err := c.Validate(&plugin); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err.Error(),
		})
	}

	updated, err := s.db.UpdatePlugin(c.Request().Context(), pluginID, plugin)
	if err != nil {
		message := echo.Map{
			"message": "failed to update plugin",
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.JSON(http.StatusOK, updated)
}

func (s *Server) DeletePlugin(c echo.Context) error {
	pluginID := c.Param("pluginId")
	if pluginID == "" {
		message := echo.Map{
			"message": "failed to delete plugin",
			"error":   "plugin id is required",
		}
		return c.JSON(http.StatusBadRequest, message)
	}

	err := s.db.DeletePluginById(c.Request().Context(), pluginID)
	if err != nil {
		message := echo.Map{
			"message": "failed to delete plugin",
			"error":   err.Error(),
		}
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, message)
	}

	return c.NoContent(http.StatusNoContent)
}

func (s *Server) AttachPluginTag(c echo.Context) error {
	pluginID := c.Param("pluginId")
	if pluginID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "failed to find plugin",
			"error":   "plugin id is required",
		})
	}
	plugin, err := s.db.FindPluginById(c.Request().Context(), pluginID)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "failed to find plugin",
			"error":   "plugin not found",
		})
	}

	var createTagDto types.CreateTagDto
	if err := c.Bind(&createTagDto); err != nil {
		return fmt.Errorf("fail to parse request, err: %w", err)
	}
	if err := c.Validate(&createTagDto); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": err.Error(),
		})
	}

	var tag *types.Tag
	tag, err = s.db.FindTagByName(c.Request().Context(), createTagDto.Name)
	if err != nil {
		if err.Error() == "no rows in result set" {
			tag, err = s.db.CreateTag(c.Request().Context(), createTagDto)
			if err != nil {
				s.logger.Error(err)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"message": "failed to create tag",
				})
			}
		} else {
			s.logger.Error(err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "failed to check for existing tag",
			})
		}
	}

	updatedPlugin, err := s.db.AttachTagToPlugin(c.Request().Context(), plugin.ID, tag.ID)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to attach tag",
		})
	}

	return c.JSON(http.StatusOK, updatedPlugin)
}

func (s *Server) DetachPluginTag(c echo.Context) error {
	pluginID := c.Param("pluginId")
	if pluginID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "failed to find plugin",
			"error":   "plugin id is required",
		})
	}
	plugin, err := s.db.FindPluginById(c.Request().Context(), pluginID)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "failed to find plugin",
			"error":   "plugin not found",
		})
	}

	tagID := c.Param("tagId")
	if tagID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "failed to find tag",
			"error":   "tag id is required",
		})
	}
	tag, err := s.db.FindTagById(c.Request().Context(), tagID)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "failed to find tag",
			"error":   "tag not found",
		})
	}

	updatedPlugin, err := s.db.DetachTagFromPlugin(c.Request().Context(), plugin.ID, tag.ID)
	if err != nil {
		s.logger.Error(err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to detach tag",
		})
	}

	return c.JSON(http.StatusOK, updatedPlugin)
}

func (s *Server) verifyPolicySignature(policy types.PluginPolicy, update bool) bool {
	msgHex, err := policyToMessageHex(policy, update)
	if err != nil {
		s.logger.Error(fmt.Errorf("failed to convert policy to message hex: %w", err))
		return false
	}

	msgBytes, err := hex.DecodeString(strings.TrimPrefix(msgHex, "0x"))
	if err != nil {
		s.logger.Error(fmt.Errorf("failed to decode message bytes: %w", err))
		return false
	}

	signatureBytes, err := hex.DecodeString(strings.TrimPrefix(policy.Signature, "0x"))
	if err != nil {
		s.logger.Error(fmt.Errorf("failed to decode signature bytes: %w", err))
		return false
	}

	isVerified, err := sigutil.VerifySignature(policy.PublicKey, policy.ChainCodeHex, policy.DerivePath, msgBytes, signatureBytes)
	if err != nil {
		s.logger.Error(fmt.Errorf("failed to verify signature: %w", err))
		return false
	}
	return isVerified
}

func policyToMessageHex(policy types.PluginPolicy, isUpdate bool) (string, error) {
	if !isUpdate {
		policy.ID = ""
	}
	// signature is not part of the message that is signed
	policy.Signature = ""

	serializedPolicy, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("failed to serialize policy")
	}
	return hex.EncodeToString(serializedPolicy), nil
}

func calculateTransactionHash(txData string) (string, error) {
	tx := &gtypes.Transaction{}
	rawTx, err := hex.DecodeString(txData)
	if err != nil {
		return "", fmt.Errorf("invalid transaction hex: %w", err)
	}

	err = tx.UnmarshalBinary(rawTx)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	chainID := tx.ChainId()
	signer := gtypes.NewEIP155Signer(chainID)
	hash := signer.Hash(tx).String()[2:]
	return hash, nil
}
