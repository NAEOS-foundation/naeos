// Copyright 2025 NAEOS contributors
// Live bridge execution requires replay protection before consequential side effects.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/evidence"
	"github.com/NAEOS-foundation/naeos/internal/governance/control"
	"github.com/NAEOS-foundation/naeos/internal/runtime/gateway"
)

type filesystemObserver struct{ root string }

func (o filesystemObserver) Observe(req gateway.ToolRequest, result gateway.ExecutionResult) (gateway.Observation, error) {
	obs := gateway.Observation{RequestID: req.RequestID, InvocationID: req.InvocationID, InvocationDigest: gateway.InvocationDigest(req), Status: "unavailable", Timestamp: result.Timestamp}
	// This observer currently verifies only filesystem writes by reading the
	// target artifact back. Do not claim independent observation for tools or
	// actions whose effects this observer cannot inspect.
	if req.Tool != "filesystem" || req.Action != "write" {
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
		gateway.WithRequiredRevalidation(true),
	)
	adapter := gateway.JSONToolAdapter{}
	gw.RegisterAdapter(adapter.Name(), adapter)
	// This bridge currently has an independently verifiable observer only for
	// filesystem writes. Deny unsupported tools/actions at the adapter boundary,
	// before the sandbox, rather than execute them and fail evidence afterward.
	_ = gw.GrantAdapterPolicy(adapter.Name(), gateway.AdapterPolicy{
		AllowedTools:   []string{"filesystem"},
		AllowedActions: []string{"write"},
	})
	return gw
}

func appendRuntimeEvidence(path string, data []byte) error {
	if path == "" {
		return fmt.Errorf("runtime evidence path is required")
	}
	if err := os.MkdirAll(filepath.Dir(filepath.Clean(path)), 0o750); err != nil {
		return fmt.Errorf("create runtime evidence directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open runtime evidence file: %w", err)
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		_ = file.Close()
		return fmt.Errorf("append runtime evidence: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync runtime evidence: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close runtime evidence: %w", err)
	}
	return nil
}

func newRuntimeBridgeCommand() *cobra.Command {
	var root string
	var replayDB, evidenceFile, evidenceSigningKey, evidenceIssuer, evidenceKeyID string
	var evidenceLedgerPath, evidenceReceiptPath, evidenceRunID, evidenceTrustedPublicKey string
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

			durableMode := evidenceLedgerPath != "" || evidenceReceiptPath != "" || evidenceRunID != "" || evidenceTrustedPublicKey != ""
			var signingKey ed25519.PrivateKey
			var trustedPublicKey ed25519.PublicKey
			var durableLedger *evidence.DurableRuntimeEventLedger
			sequence := 1
			if durableMode {
				if evidenceLedgerPath == "" || evidenceReceiptPath == "" || evidenceRunID == "" ||
					evidenceFile == "" || evidenceSigningKey == "" || evidenceTrustedPublicKey == "" ||
					evidenceIssuer == "" || evidenceKeyID == "" {
					return fmt.Errorf("authenticated durable evidence requires --evidence-file, --evidence-ledger, --evidence-receipt, --evidence-run-id, --evidence-signing-key, --evidence-trusted-public-key, --evidence-issuer, and --evidence-key-id")
				}
				keyData, err := os.ReadFile(evidenceSigningKey)
				if err != nil {
					return fmt.Errorf("read evidence signing key: %w", err)
				}
				privateKey, err := gateway.ParseEd25519PrivateKey(keyData)
				if err != nil {
					return fmt.Errorf("parse evidence signing key: %w", err)
				}
				signingKey = privateKey
				trustData, err := os.ReadFile(evidenceTrustedPublicKey)
				if err != nil {
					return fmt.Errorf("read trusted evidence public key: %w", err)
				}
				publicKey, err := evidence.ParseEd25519PublicKey(trustData)
				if err != nil {
					return fmt.Errorf("parse trusted evidence public key: %w", err)
				}
				trustedPublicKey = publicKey

				durableLedger, err = evidence.NewDurableRuntimeEventLedger(evidenceLedgerPath)
				if err != nil {
					return err
				}
				if err := durableLedger.Verify(); err != nil {
					return fmt.Errorf("existing durable runtime ledger failed verification: %w", err)
				}
				records := durableLedger.Records()
				for _, record := range records {
					if record.RunID != evidenceRunID {
						return fmt.Errorf("existing durable runtime ledger belongs to another run")
					}
				}
				sequence = len(records) + 1
				if _, err := os.Stat(evidenceReceiptPath); err == nil {
					durableLedger.Seal()
					data, readErr := os.ReadFile(evidenceReceiptPath)
					if readErr != nil {
						return readErr
					}
					var signed evidence.SignedDurableRuntimeReceipt
					if err := json.Unmarshal(data, &signed); err != nil {
						return fmt.Errorf("decode existing durable runtime receipt: %w", err)
					}
					if err := evidence.VerifySignedDurableRuntimeReceipt(signed, durableLedger, evidenceRunID, trustedPublicKey, evidenceIssuer, evidenceKeyID); err != nil {
						return fmt.Errorf("existing durable runtime receipt failed trusted verification: %w", err)
					}
					return fmt.Errorf("runtime evidence run %q is already sealed and verified", evidenceRunID)
				} else if !os.IsNotExist(err) {
					return fmt.Errorf("inspect durable runtime receipt: %w", err)
				}
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
					runtimeEvidence, evidenceErr := gateway.BuildRuntimeEvidence(result)
					if evidenceErr != nil {
						if durableMode {
							return fmt.Errorf("build required runtime evidence: %w", evidenceErr)
						}
						response["evidence_error"] = evidenceErr.Error()
					} else {
						if evidenceSigningKey != "" {
							var privateKey []byte
							if durableMode {
								privateKey = signingKey
							} else {
								keyData, keyReadErr := os.ReadFile(evidenceSigningKey)
								if keyReadErr != nil {
									return keyReadErr
								}
								parsed, keyParseErr := gateway.ParseEd25519PrivateKey(keyData)
								if keyParseErr != nil {
									return fmt.Errorf("parse evidence signing key: %w", keyParseErr)
								}
								privateKey = parsed
							}
							if signErr := gateway.SignRuntimeEvidence(&runtimeEvidence, privateKey, evidenceIssuer, evidenceKeyID); signErr != nil {
								return fmt.Errorf("sign runtime evidence: %w", signErr)
							}
						}
						evidenceData, marshalErr := json.Marshal(runtimeEvidence)
						if marshalErr != nil {
							return marshalErr
						}
						if durableMode {
							if err := appendRuntimeEvidence(evidenceFile, evidenceData); err != nil {
								return err
							}
							if _, err := durableLedger.Publish(evidenceRunID, "runtime.execution_evidence", runtimeEvidence.EvidenceDigest, sequence); err != nil {
								return fmt.Errorf("persist durable runtime evidence event: %w", err)
							}
							sequence++
						} else {
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
				}
				data, encErr := json.Marshal(response)
				if encErr != nil {
					return encErr
				}
				if _, err := io.WriteString(out, string(data)+"\n"); err != nil {
					return err
				}
			}
			if err := in.Err(); err != nil {
				return err
			}
			if durableMode {
				durableLedger.Seal()
				receipt, err := evidence.CreateDurableRuntimeReceipt(durableLedger, evidenceRunID, time.Now().UTC())
				if err != nil {
					return fmt.Errorf("create durable runtime receipt: %w", err)
				}
				privateKey := signingKey
				signedReceipt, err := evidence.SignDurableRuntimeReceipt(receipt, privateKey, evidenceIssuer, evidenceKeyID)
				if err != nil {
					return fmt.Errorf("sign durable runtime receipt: %w", err)
				}
				if err := evidence.VerifySignedDurableRuntimeReceipt(signedReceipt, durableLedger, evidenceRunID, trustedPublicKey, evidenceIssuer, evidenceKeyID); err != nil {
					return fmt.Errorf("verify signed durable runtime receipt against configured trust anchor: %w", err)
				}
				receiptData, err := json.MarshalIndent(signedReceipt, "", "  ")
				if err != nil {
					return err
				}
				if err := writeAtomicRuntimeReceipt(evidenceReceiptPath, append(receiptData, '\n')); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "filesystem-root", "", "filesystem sandbox root for governed writes")
	cmd.Flags().StringVar(&replayDB, "replay-db", "", "named database connection for durable replay protection")
	cmd.Flags().StringVar(&evidenceFile, "evidence-file", "", "append canonical runtime evidence JSONL to this file")
	cmd.Flags().StringVar(&evidenceSigningKey, "evidence-signing-key", "", "Ed25519 private key file for authenticated runtime evidence")
	cmd.Flags().StringVar(&evidenceIssuer, "evidence-issuer", "", "trusted issuer identity embedded in runtime evidence signatures")
	cmd.Flags().StringVar(&evidenceKeyID, "evidence-key-id", "", "trusted signing key identifier embedded in runtime evidence signatures")
	cmd.Flags().StringVar(&evidenceLedgerPath, "evidence-ledger", "", "append-only durable runtime event ledger for authenticated evidence")
	cmd.Flags().StringVar(&evidenceReceiptPath, "evidence-receipt", "", "signed completion receipt written after the durable ledger is sealed")
	cmd.Flags().StringVar(&evidenceRunID, "evidence-run-id", "", "stable run identity bound to all durable runtime evidence events")
	cmd.Flags().StringVar(&evidenceTrustedPublicKey, "evidence-trusted-public-key", "", "independently trusted Ed25519 public key used to verify the completion receipt")
	return cmd
}

func writeAtomicRuntimeReceipt(path string, data []byte) error {
	if path == "" {
		return fmt.Errorf("runtime receipt path is required")
	}
	if err := os.MkdirAll(filepath.Dir(filepath.Clean(path)), 0o750); err != nil {
		return fmt.Errorf("create runtime receipt directory: %w", err)
	}
	tmp := filepath.Clean(path) + ".tmp"
	file, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create runtime receipt temporary file: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write runtime receipt: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync runtime receipt: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close runtime receipt: %w", err)
	}
	if err := os.Rename(tmp, filepath.Clean(path)); err != nil {
		return fmt.Errorf("commit runtime receipt: %w", err)
	}
	return nil
}
