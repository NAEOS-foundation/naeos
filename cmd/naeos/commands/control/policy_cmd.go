// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package control

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	clihelpers "github.com/NAEOS-foundation/naeos/cmd/naeos/internal/clihelpers"
	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
	"github.com/NAEOS-foundation/naeos/internal/governance/policy"
)

// saveVersions persists every version of a policy to a single file,
// preserving each version's Active flag exactly as carried by the registry.
func saveVersions(reg *policy.Registry, id string) error {
	path, err := clihelpers.PolicyPath(id)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "creating policies directory")
	}
	versions := reg.Versions(id)
	if len(versions) == 0 {
		return nil
	}
	data, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "marshaling policy")
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return naeoserr.Wrapf(err, naeoserr.ErrInternal, "writing policy file")
	}
	return nil
}

func NewPolicyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage governance policies",
		Long:  `Register, list, and validate NAEOS governance policies used by the control plane to authorize agent actions.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(NewPolicyRegisterCommand())
	cmd.AddCommand(NewPolicyListCommand())
	cmd.AddCommand(NewPolicyGetCommand())
	cmd.AddCommand(NewPolicySetActiveCommand())
	cmd.AddCommand(NewPolicyValidateCommand())
	return cmd
}

func NewPolicyRegisterCommand() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a governance policy from a JSON file",
		Long: `Register a governance policy, persisting it for use by the control plane.

Example:
  naeos policy register --file policy.json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}
			data, err := os.ReadFile(file)
			if err != nil {
				return naeoserr.Wrapf(err, naeoserr.ErrInternal, "reading policy file")
			}
			var p policy.Policy
			if err := json.Unmarshal(data, &p); err != nil {
				return naeoserr.Wrapf(err, naeoserr.ErrValidation, "parsing policy JSON")
			}
			if err := policy.Validate(&p); err != nil {
				return err
			}
			reg, err := clihelpers.LoadPolicies()
			if err != nil {
				return err
			}
			if err := reg.Register(&p); err != nil {
				return err
			}
			if err := saveVersions(reg, p.ID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Policy %s v%s registered\n", p.ID, p.Version)
			return nil
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to policy JSON file (required)")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func NewPolicyListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registered governance policies",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := clihelpers.LoadPolicies()
			if err != nil {
				return err
			}
			policies := reg.List()
			sort.Slice(policies, func(i, j int) bool { return policies[i].ID < policies[j].ID })
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-20s %-10s %-30s %s\n", "POLICY", "VERSION", "SCOPE", "DEFAULT")
			fmt.Fprintf(out, "%-20s %-10s %-30s %s\n", strings.Repeat("-", 20), strings.Repeat("-", 10), strings.Repeat("-", 30), strings.Repeat("-", 7))
			for _, p := range policies {
				scope := fmt.Sprintf("%s/%s (%s)", orWildcard(p.Scope.Resource), orWildcard(p.Scope.Action), orWildcard(p.Scope.Environment))
				fmt.Fprintf(out, "%-20s %-10s %-30s %s\n", p.ID, p.Version, scope, p.Default)
			}
			return nil
		},
	}
}

func orWildcard(s string) string {
	if s == "" {
		return "*"
	}
	return s
}

func NewPolicyGetCommand() *cobra.Command {
	var version string
	cmd := &cobra.Command{
		Use:   "get <policy-id>",
		Short: "Show a registered governance policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := clihelpers.LoadPolicies()
			if err != nil {
				return err
			}
			var p *policy.Policy
			var ok bool
			if version != "" {
				p, ok = reg.Get(args[0], version)
			} else {
				p, ok = reg.GetActive(args[0])
			}
			if !ok {
				return naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("policy %s not found", args[0]))
			}
			data, err := json.MarshalIndent(p, "", "  ")
			if err != nil {
				return naeoserr.Wrapf(err, naeoserr.ErrInternal, "marshaling policy")
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
	cmd.Flags().StringVar(&version, "version", "", "specific policy version to show")
	return cmd
}

func NewPolicySetActiveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-active <policy-id> <version>",
		Short: "Set the active version of a governance policy",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			reg, err := clihelpers.LoadPolicies()
			if err != nil {
				return err
			}
			if err := reg.SetActive(args[0], args[1]); err != nil {
				return err
			}
			if err := saveVersions(reg, args[0]); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Policy %s active version set to %s\n", args[0], args[1])
			return nil
		},
	}
	return cmd
}

func NewPolicyValidateCommand() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a governance policy file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}
			data, err := os.ReadFile(file)
			if err != nil {
				return naeoserr.Wrapf(err, naeoserr.ErrInternal, "reading policy file")
			}
			var p policy.Policy
			if err := json.Unmarshal(data, &p); err != nil {
				return naeoserr.Wrapf(err, naeoserr.ErrValidation, "parsing policy JSON")
			}
			if err := policy.Validate(&p); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Policy %s v%s is valid\n", p.ID, p.Version)
			return nil
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "path to policy JSON file (required)")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}
