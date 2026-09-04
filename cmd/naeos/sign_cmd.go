package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/signing"
)

func newSignCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "Digital artifact signing and verification",
		Long:  `Sign artifacts with Ed25519 keys, verify signatures, and manage key pairs.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newSignKeygenCommand())
	cmd.AddCommand(newSignSignCommand())
	cmd.AddCommand(newSignVerifyCommand())
	return cmd
}

func newSignKeygenCommand() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate a new Ed25519 key pair",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			kp, err := signing.GenerateKeyPair()
			if err != nil {
				return err
			}

			keys := map[string]string{
				"publicKey":  kp.PublicKeyBase64(),
				"privateKey": kp.PrivateKeyBase64(),
			}
			data, _ := json.MarshalIndent(keys, "", "  ")

			if outputPath != "" {
				data, _ := json.MarshalIndent(keys, "", "  ")
				if err := os.WriteFile(outputPath, data, 0o600); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Key pair written to %s\n", outputPath)
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "write key pair to file (JSON)")
	return cmd
}

func newSignSignCommand() *cobra.Command {
	var (
		keyPath  string
		output   string
		signer   string
		outputFmt string
	)

	cmd := &cobra.Command{
		Use:   "sign <artifact-path>",
		Short: "Sign an artifact file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			kp, err := loadKeyPair(keyPath)
			if err != nil {
				return err
			}

			hash, size, err := signing.HashFile(path)
			if err != nil {
				return err
			}
			_ = hash

			content, err := readFile(path)
			if err != nil {
				return err
			}

			bundle, err := signing.Sign(path, content, kp,
				signing.WithSigner(signer),
				signing.WithPath(path),
				signing.WithSize(size),
			)
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(bundle, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			sigPath := output
			if sigPath == "" {
				sigPath = path + ".sig.json"
			}
			if err := signing.Write(bundle, sigPath); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Signature written to %s\n", sigPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&keyPath, "key", "", "path to private key file (JSON)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output signature path (default: <artifact>.sig.json)")
	cmd.Flags().StringVar(&signer, "signer", "", "signer identity")
	cmd.Flags().StringVar(&outputFmt, "output-format", "file", "output format: file or json")
	return cmd
}

func newSignVerifyCommand() *cobra.Command {
	var outputFmt string

	cmd := &cobra.Command{
		Use:   "verify <signature-path>",
		Short: "Verify an artifact signature",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			result, err := signing.VerifyFile(path)
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(data))
				return nil
			}

			out := cmd.OutOrStdout()
			mark := "PASS"
			if !result.Valid {
				mark = "FAIL"
			}
			fmt.Fprintf(out, "[%s] Signature %s\n", mark, result.Message)
			fmt.Fprintf(out, "  Algorithm:   %s\n", result.Algorithm)
			fmt.Fprintf(out, "  Artifact:    %s\n", result.ArtifactName)
			fmt.Fprintf(out, "  Hash:        %s\n", result.ArtifactHash)
			fmt.Fprintf(out, "  Public Key:  %s...%s\n",
				result.PublicKey[:min(8, len(result.PublicKey))],
				result.PublicKey[max(0, len(result.PublicKey)-8):])
			return nil
		},
	}

	cmd.Flags().StringVar(&outputFmt, "output", "table", "output format: table or json")
	return cmd
}

func loadKeyPair(path string) (*signing.KeyPair, error) {
	if path == "" {
		return nil, fmt.Errorf("--key is required")
	}
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}
	var keys map[string]string
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, fmt.Errorf("parse key file: %w", err)
	}
	privKey, err := signing.ParsePrivateKeyBase64(keys["privateKey"])
	if err != nil {
		return nil, err
	}
	pubKey, err := signing.ParsePublicKeyBase64(keys["publicKey"])
	if err != nil {
		return nil, err
	}
	return &signing.KeyPair{PublicKey: pubKey, PrivateKey: privKey}, nil
}

func readFile(path string) ([]byte, error) {
	return readFileImpl(path)
}

func readFileImpl(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return data, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
