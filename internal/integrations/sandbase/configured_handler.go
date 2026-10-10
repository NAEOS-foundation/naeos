// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package sandbase

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/NAEOS-foundation/naeos/internal/controlplane"
)

// RuntimeConfig is the operator-managed policy/grant input for the optional
// SandBase endpoint. Keep the bearer token outside this file in a secret env var.
type RuntimeConfig struct {
	Policy controlplane.Policy `json:"policy"`
	Grant  controlplane.Grant  `json:"grant"`
}

type configuredHandler struct {
	configPath string
	token      string
	store      *controlplane.PolicyStore
	evaluator  *controlplane.Evaluator
	ledger     *controlplane.Ledger

	mu           sync.Mutex
	lastVersion  int
	lastDigest   string
	policyID     string
}

func NewConfiguredHTTPHandler(configPath, bearerToken, ledgerPath string) (http.Handler, error) {
	if strings.TrimSpace(configPath) == "" {
		return nil, fmt.Errorf("SandBase authorization config path is required")
	}
	if strings.TrimSpace(bearerToken) == "" {
		return nil, fmt.Errorf("SandBase authorization bearer token is required")
	}
	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("resolve SandBase authorization config path: %w", err)
	}
	if ledgerPath == "" {
		ledgerPath = absConfigPath + ".ledger.json"
	}
	absLedgerPath, err := filepath.Abs(ledgerPath)
	if err != nil {
		return nil, fmt.Errorf("resolve SandBase authorization ledger path: %w", err)
	}

	ledger, err := loadOrCreateLedger(absLedgerPath)
	if err != nil {
		return nil, err
	}
	ledger.SetPersistencePath(absLedgerPath)

	store := controlplane.NewPolicyStore()
	h := &configuredHandler{
		configPath: absConfigPath,
		token:      bearerToken,
		store:      store,
		evaluator:  controlplane.NewEvaluator(store),
		ledger:     ledger,
	}
	if _, _, err := h.refreshPolicy(); err != nil {
		return nil, fmt.Errorf("load initial SandBase authorization config: %w", err)
	}
	return h, nil
}

func loadOrCreateLedger(path string) (*controlplane.Ledger, error) {
	if _, err := os.Stat(path); err == nil {
		ledger, loadErr := controlplane.LoadLedger(path)
		if loadErr != nil {
			return nil, fmt.Errorf("load SandBase authorization ledger: %w", loadErr)
		}
		return ledger, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("inspect SandBase authorization ledger: %w", err)
	}
	return controlplane.NewLedger(), nil
}

func (h *configuredHandler) refreshPolicy() (controlplane.Policy, controlplane.Grant, error) {
	data, err := os.ReadFile(h.configPath)
	if err != nil {
		return controlplane.Policy{}, controlplane.Grant{}, fmt.Errorf("read config: %w", err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	var config RuntimeConfig
	if err := decoder.Decode(&config); err != nil {
		return controlplane.Policy{}, controlplane.Grant{}, fmt.Errorf("decode config: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return controlplane.Policy{}, controlplane.Grant{}, fmt.Errorf("config must contain exactly one JSON value")
	}
	if err := validateRuntimeConfig(config); err != nil {
		return controlplane.Policy{}, controlplane.Grant{}, err
	}
	canonical, err := json.Marshal(config.Policy)
	if err != nil {
		return controlplane.Policy{}, controlplane.Grant{}, fmt.Errorf("marshal policy: %w", err)
	}
	sum := sha256.Sum256(canonical)
	digest := hex.EncodeToString(sum[:])

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.policyID != "" && h.policyID != config.Policy.ID {
		return controlplane.Policy{}, controlplane.Grant{}, fmt.Errorf("policy ID cannot change while the handler is running")
	}
	if h.lastVersion > 0 && config.Policy.Version < h.lastVersion {
		return controlplane.Policy{}, controlplane.Grant{}, fmt.Errorf("policy version rollback is not allowed")
	}
	if h.lastVersion == config.Policy.Version && h.lastDigest != digest {
		return controlplane.Policy{}, controlplane.Grant{}, fmt.Errorf("policy contents changed without incrementing policy version")
	}
	if err := h.store.Set(&config.Policy); err != nil {
		return controlplane.Policy{}, controlplane.Grant{}, fmt.Errorf("activate policy: %w", err)
	}
	h.policyID = config.Policy.ID
	h.lastVersion = config.Policy.Version
	h.lastDigest = digest
	return config.Policy, config.Grant, nil
}

func validateRuntimeConfig(config RuntimeConfig) error {
	policy, grant := config.Policy, config.Grant
	if strings.TrimSpace(policy.ID) == "" || policy.Version < 1 || policy.Status != "active" {
		return fmt.Errorf("policy must have an ID, positive version, and active status")
	}
	if strings.TrimSpace(grant.GrantID) == "" || strings.TrimSpace(grant.AgentID) == "" ||
		grant.PolicyID != policy.ID || grant.PolicyVersion < 1 || grant.Status != "active" ||
		grant.Revoked || len(grant.Capabilities) == 0 {
		return fmt.Errorf("grant must be active, unrevoked, scoped to an agent/policy, and include capabilities")
	}
	return nil
}

func (h *configuredHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	policy, grant, err := h.refreshPolicy()
	if err != nil {
		writeSandBaseDecision(w, http.StatusServiceUnavailable, SandBaseAuthorizationResponse{
			Decision: "deny", Reason: "authorization_unavailable",
		})
		return
	}
	adapter := &Adapter{
		Gateway: controlplane.NewDecisionGateway(h.evaluator, h.ledger),
		Policy:  &policy,
		Grant:   &grant,
	}
	NewHTTPHandler(adapter, h.token).ServeHTTP(w, r)
}
