// Copyright 2025 NAEOS contributors
// Live bridge execution requires replay protection before consequential side effects.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/runtime/gateway"
)

type filesystemObserver struct{ root string }

func (o filesystemObserver) Observe(req gateway.ToolRequest, result gateway.ExecutionResult) (gateway.Observation, error) {
	obs := gateway.Observation{RequestID: req.RequestID, InvocationID: req.InvocationID, InvocationDigest: gateway.InvocationDigest(req), Status: "absent", Timestamp: result.Timestamp}
	if req.Tool != "filesystem" || req.Action != "write" {
		obs.Status = "observed"
		obs.Observed = true
		return obs, nil
	}
	rel, ok := req.Payload["path"].(string)
	if !ok || rel == "" {
		return obs, fmt.Errorf("filesystem observer requires payload.path")
	}
	root, err := filepath.Abs(o.root)
	if err != nil {
		return obs, err
	}
	target, err := filepath.Abs(filepath.Join(root, rel))
	if err != nil {
		return obs, err
	}
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || len(relative) > 3 && relative[:3] == "../" {
		return obs, fmt.Errorf("observed path escapes filesystem root")
	}
	data, err := os.ReadFile(target)
	if os.IsNotExist(err) {
		return obs, nil
	}
	if err != nil {
		obs.Status = "unavailable"
		return obs, err
	}
	h := sha256.Sum256(data)
	obs.Status = "observed"
	obs.Observed = true
	obs.ArtifactHash = "sha256:" + hex.EncodeToString(h[:])
	obs.ArtifactSize = int64(len(data))
	obs.Metadata = map[string]string{"path": rel}
	return obs, nil
}

func newRuntimeBridgeGateway(cp gateway.ControlPlane, root string, replayStore gateway.InvocationStore) *gateway.ExecutionGateway {
	if replayStore == nil {
		replayStore = gateway.NewInMemoryInvocationStore()
	}
	sb := gateway.NewDefaultSandbox(gateway.SandboxConfig{FilesystemRoot: root})
	gw := gateway.New(
		cp,
		sb,
		gateway.WithObserver(filesystemObserver{root: root}),
		gateway.WithReplayProtection(true),
		gateway.WithInvocationStore(replayStore),
	)
	adapter := gateway.JSONToolAdapter{}
	gw.RegisterAdapter(adapter.Name(), adapter)
	_ = gw.GrantAdapterPolicy(adapter.Name(), gateway.AdapterPolicy{
		AllowedTools:   []string{"*"},
		AllowedActions: []string{"*"},
	})
	return gw
}

func newRuntimeBridgeCommand() *cobra.Command {
	var root string
	var replayDB, evidenceFile, evidenceSigningKey, evidenceIssuer, evidenceKeyID string
	cmd := &cobra.Command{
		Use:   "bridge",
		Short: "Stream JSON tool calls through the governance gateway",
		Long:  "Read one JSON tool request per line from stdin and emit one JSON execution result per line. This is the protocol-neutral live agent interception boundary.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := loadPolicies()
			if err != nil {
				return err
			}
			if root == "" {
				root = os.TempDir()
			}
			cp := control.New(reg)
			replayStore, closeReplayStore, err := loadReplayStore(replayDB)
			if err != nil {
				return err
			}
			defer closeReplayStore()
			gw := newRuntimeBridgeGateway(cp, root, replayStore)

			in := bufio.NewScanner(cmd.InOrStdin())
			out := cmd.OutOrStdout()
			for in.Scan() {
				var raw map[string]any
				if err := json.Unmarshal(in.Bytes(), &raw); err != nil {
					return fmt.Errorf("invalid bridge request: %w", err)
				}
				result, err := gw.AuthorizeFromAdapter("json", raw)
				response := map[string]any{"result": result}
				if err != nil {
					response["error"] = err.Error()
				}
				if evidenceFile != "" {
					evidence, evidenceErr := gateway.BuildRuntimeEvidence(result)
					if evidenceErr != nil {
						response["evidence_error"] = evidenceErr.Error()
					} else {
						if evidenceSigningKey != "" {
							keyData, keyReadErr := os.ReadFile(evidenceSigningKey)
							if keyReadErr != nil {
								return keyReadErr
							}
							privateKey, keyParseErr := gateway.ParseEd25519PrivateKey(keyData)
							if keyParseErr != nil {
								return fmt.Errorf("parse evidence signing key: %w", keyParseErr)
							}
							if signErr := gateway.SignRuntimeEvidence(&evidence, privateKey, evidenceIssuer, evidenceKeyID); signErr != nil {
								return fmt.Errorf("sign runtime evidence: %w", signErr)
							}
						}
						evidenceData, evidenceMarshalErr := json.Marshal(evidence)
						if evidenceMarshalErr != nil {
							return evidenceMarshalErr
						}
						file, openErr := os.OpenFile(evidenceFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
						if openErr != nil {
							return openErr
						}
						if _, writeErr := file.Write(append(evidenceData, '\n')); writeErr != nil {
							_ = file.Close()
							return writeErr
						}
						if closeErr := file.Close(); closeErr != nil {
							return closeErr
						}
					}
				}
				data, encErr := json.Marshal(response)
				if encErr != nil {
					return encErr
				}
				if _, err := io.WriteString(out, string(data)+"\n"); err != nil {
					return err
				}
			}
			return in.Err()
		},
	}
	cmd.Flags().StringVar(&root, "filesystem-root", "", "filesystem sandbox root for governed writes")
	cmd.Flags().StringVar(&replayDB, "replay-db", "", "named database connection for durable replay protection")
	cmd.Flags().StringVar(&evidenceFile, "evidence-file", "", "append canonical runtime evidence JSONL to this file")
	cmd.Flags().StringVar(&evidenceSigningKey, "evidence-signing-key", "", "Ed25519 private key file for authenticated runtime evidence")
	cmd.Flags().StringVar(&evidenceIssuer, "evidence-issuer", "", "trusted issuer identity embedded in runtime evidence signatures")
	cmd.Flags().StringVar(&evidenceKeyID, "evidence-key-id", "", "trusted signing key identifier embedded in runtime evidence signatures")
	return cmd
}
