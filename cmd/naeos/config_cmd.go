package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/configprovider"
)

func newConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Resolve configuration from environment, files, secrets, and Vault",
		Long:  `Use references (env:VAR, file:/path, secret:ns/name/key, vault:path#key) to resolve configuration values.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newConfigResolveCommand())
	cmd.AddCommand(newConfigTestCommand())
	cmd.AddCommand(newConfigSourcesCommand())
	return cmd
}

func buildChain(secrets, vault []string) (*configprovider.Chain, error) {
	chain := configprovider.NewChain(configprovider.NewEnvProvider(), configprovider.NewFileProvider())

	if len(secrets) > 0 {
		sec := configprovider.NewK8sSecretProvider()
		for _, s := range secrets {
			loc, val, err := splitPair(s, "namespace/name/key")
			if err != nil {
				return nil, err
			}
			parts := strings.Split(loc, "/")
			if len(parts) != 3 {
				return nil, fmt.Errorf("invalid secret %q (want namespace/name/key)", loc)
			}
			sec.AddSecret(parts[0], parts[1], parts[2], val)
		}
		chain.Add(sec)
	}

	if len(vault) > 0 {
		vp := configprovider.NewVaultProvider()
		for _, s := range vault {
			loc, val, err := splitPair(s, "path#key")
			if err != nil {
				return nil, err
			}
			parts := strings.Split(loc, "#")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid vault %q (want path#key)", loc)
			}
			vp.AddSecret(parts[0], parts[1], val)
		}
		chain.Add(vp)
	}
	return chain, nil
}

func splitPair(pair, format string) (string, string, error) {
	idx := strings.Index(pair, "=")
	if idx < 0 {
		return "", "", fmt.Errorf("invalid value %q (want %s=value)", pair, format)
	}
	return pair[:idx], pair[idx+1:], nil
}

func newConfigResolveCommand() *cobra.Command {
	var (
		outputFmt string
		secrets   []string
		vault     []string
	)

	cmd := &cobra.Command{
		Use:   "resolve <config.json>",
		Short: "Resolve a configuration file into concrete values",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chain, err := buildChain(secrets, vault)
			if err != nil {
				return err
			}

			data, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			var input map[string]string
			if err := json.Unmarshal(data, &input); err != nil {
				return err
			}

			resolved, err := configprovider.NewResolver(chain).ResolveMap(input)
			if err != nil {
				return err
			}

			if outputFmt == "json" {
				out, _ := json.MarshalIndent(resolved, "", "  ")
				fmt.Fprintln(cmd.OutOrStdout(), string(out))
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-24s %-28s %-10s %s\n", "KEY", "VALUE", "SOURCE", "REFERENCE")
			for _, cfg := range resolved {
				fmt.Fprintf(out, "%-24s %-28q %-10s\n", cfg.Key, cfg.Value, cfg.Provider)
			}
			fmt.Fprintf(out, "resolved %d keys from %s\n", len(resolved), args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&outputFmt, "output", "table", "output format: table or json")
	cmd.Flags().StringSliceVarP(&secrets, "secret", "s", nil, "inject secret ns/name/key=value (repeatable)")
	cmd.Flags().StringSliceVarP(&vault, "vault", "v", nil, "inject vault path#key=value (repeatable)")
	return cmd
}

func newConfigTestCommand() *cobra.Command {
	var (
		secrets []string
		vault   []string
	)

	cmd := &cobra.Command{
		Use:   "test <reference>",
		Short: "Resolve a single config reference (e.g. env:FOO)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chain, err := buildChain(secrets, vault)
			if err != nil {
				return err
			}
			ref, err := configprovider.ParseReference(args[0])
			if err != nil {
				return err
			}
			res, err := chain.Resolve(ref)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Provider: %s\nScheme:   %s\nValue:    %q\n", res.Provider, res.Scheme, res.Value)
			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&secrets, "secret", "s", nil, "inject secret ns/name/key=value (repeatable)")
	cmd.Flags().StringSliceVarP(&vault, "vault", "v", nil, "inject vault path#key=value (repeatable)")
	return cmd
}

func newConfigSourcesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sources",
		Short: "List available config sources and their reference syntax",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "SOURCES")
			fmt.Fprintln(out, "  env     env:VAR_NAME          environment variable")
			fmt.Fprintln(out, "  file    file:/path/to/file   trimmed file contents")
			fmt.Fprintln(out, "  secret  secret:ns/name/key   Kubernetes secret (--secret)")
			fmt.Fprintln(out, "  vault   vault:path#key       HashiCorp Vault KV (--vault)")
			fmt.Fprintln(out, "  plain   value                passed through verbatim")
			return nil
		},
	}
	return cmd
}