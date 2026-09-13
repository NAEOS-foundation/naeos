// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/agent"
)

func newAgentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage autonomous agent sessions",
		Long:  `Create and inspect agent sessions that capture project context, permissions, and action history.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newAgentCreateCommand())
	cmd.AddCommand(newAgentListCommand())
	cmd.AddCommand(newAgentGetCommand())
	cmd.AddCommand(newAgentAppendActionCommand())
	cmd.AddCommand(newAgentListActionsCommand())
	cmd.AddCommand(newAgentDeleteCommand())
	return cmd
}

func newAgentCreateCommand() *cobra.Command {
	var storePath string
	var outputFormat string
	var agentID, project, workspace, specification, specificationVersion string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new agent session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if agentID == "" {
				return fmt.Errorf("--agent-id is required")
			}

			store, err := loadAgentStore(storePath)
			if err != nil {
				return err
			}

			session, err := store.CreateSession(agent.Session{
				AgentID:              agentID,
				Project:              project,
				Workspace:            workspace,
				Specification:        specification,
				SpecificationVersion: specificationVersion,
				Status:               "active",
			})
			if err != nil {
				return err
			}
			if err := store.Save(); err != nil {
				return err
			}

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), session, outputFormat)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created session %s for agent %s\n", session.ID, session.AgentID)
			if session.Project != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Project: %s\n", session.Project)
			}
			if session.Specification != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Specification: %s\n", session.Specification)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&storePath, "store-path", "", "path to the agent session store JSON file")
	cmd.Flags().StringVar(&agentID, "agent-id", "", "agent identifier (required)")
	cmd.Flags().StringVar(&project, "project", "", "project name for the session")
	cmd.Flags().StringVar(&workspace, "workspace", "", "workspace path for the session")
	cmd.Flags().StringVar(&specification, "specification", "", "specification reference for the session")
	cmd.Flags().StringVar(&specificationVersion, "specification-version", "", "specification version for the session")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "output format: text, json, yaml")
	_ = cmd.MarkFlagRequired("agent-id")
	return cmd
}

func newAgentListCommand() *cobra.Command {
	var storePath string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stored agent sessions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := loadAgentStore(storePath)
			if err != nil {
				return err
			}
			sessions := store.ListSessions()
			agent.SortSessionsByCreatedAtDescending(sessions)

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), sessions, outputFormat)
			}

			if len(sessions) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No agent sessions found.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", "Sessions")
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", "--------")
			for _, session := range sessions {
				fmt.Fprintf(cmd.OutOrStdout(), "%s | agent=%s | status=%s | project=%s\n",
					session.ID,
					session.AgentID,
					session.Status,
					session.Project,
				)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&storePath, "store-path", "", "path to the agent session store JSON file")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "output format: text, json, yaml")
	return cmd
}

func newAgentGetCommand() *cobra.Command {
	var storePath string
	var outputFormat string
	var sessionID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a stored agent session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sessionID == "" {
				return fmt.Errorf("--session-id is required")
			}

			store, err := loadAgentStore(storePath)
			if err != nil {
				return err
			}
			session, ok := store.GetSession(sessionID)
			if !ok {
				return fmt.Errorf("session %s not found", sessionID)
			}

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), session, outputFormat)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Session: %s\n", session.ID)
			fmt.Fprintf(cmd.OutOrStdout(), "Agent: %s\n", session.AgentID)
			fmt.Fprintf(cmd.OutOrStdout(), "Status: %s\n", session.Status)
			if session.Project != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Project: %s\n", session.Project)
			}
			if session.Workspace != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Workspace: %s\n", session.Workspace)
			}
			if session.Specification != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Specification: %s\n", session.Specification)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Actions: %d\n", len(session.Actions))
			return nil
		},
	}

	cmd.Flags().StringVar(&storePath, "store-path", "", "path to the agent session store JSON file")
	cmd.Flags().StringVar(&sessionID, "session-id", "", "session identifier (required)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "output format: text, json, yaml")
	_ = cmd.MarkFlagRequired("session-id")
	return cmd
}

func newAgentAppendActionCommand() *cobra.Command {
	var storePath string
	var outputFormat string
	var sessionID, actionType, target, agentID, reason, policyID, decision string
	var specRefs []string
	var parametersJSON string

	cmd := &cobra.Command{
		Use:   "append-action",
		Short: "Append an action to a stored session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sessionID == "" {
				return fmt.Errorf("--session-id is required")
			}
			if actionType == "" {
				return fmt.Errorf("--type is required")
			}
			if target == "" {
				return fmt.Errorf("--target is required")
			}

			store, err := loadAgentStore(storePath)
			if err != nil {
				return err
			}
			if _, ok := store.GetSession(sessionID); !ok {
				return fmt.Errorf("session %s not found", sessionID)
			}

			var params map[string]any
			if parametersJSON != "" {
				if err := json.Unmarshal([]byte(parametersJSON), &params); err != nil {
					return fmt.Errorf("parse parameters json: %w", err)
				}
			}

			action, err := store.AppendAction(sessionID, agent.Action{
				AgentID:   agentID,
				Type:      actionType,
				Target:    target,
				Reason:    reason,
				SpecRefs:  specRefs,
				Parameters: params,
				Decision:  decision,
				PolicyID:  policyID,
			})
			if err != nil {
				return err
			}
			if err := store.Save(); err != nil {
				return err
			}

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), action, outputFormat)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Appended action %s to session %s\n", action.ID, sessionID)
			fmt.Fprintf(cmd.OutOrStdout(), "Type: %s\n", action.Type)
			fmt.Fprintf(cmd.OutOrStdout(), "Target: %s\n", action.Target)
			return nil
		},
	}

	cmd.Flags().StringVar(&storePath, "store-path", "", "path to the agent session store JSON file")
	cmd.Flags().StringVar(&sessionID, "session-id", "", "session identifier (required)")
	cmd.Flags().StringVar(&actionType, "type", "", "action type (required)")
	cmd.Flags().StringVar(&target, "target", "", "target resource or tool (required)")
	cmd.Flags().StringVar(&agentID, "agent-id", "", "agent identifier for the action")
	cmd.Flags().StringVar(&reason, "reason", "", "reason for the action")
	cmd.Flags().StringVar(&policyID, "policy-id", "", "policy identifier that authorized the action")
	cmd.Flags().StringVar(&decision, "decision", "", "decision outcome for the action")
	cmd.Flags().StringArrayVar(&specRefs, "spec-ref", nil, "specification reference to attach to the action")
	cmd.Flags().StringVar(&parametersJSON, "parameters", "", "JSON object of additional parameters")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "output format: text, json, yaml")
	_ = cmd.MarkFlagRequired("session-id")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}

func newAgentListActionsCommand() *cobra.Command {
	var storePath string
	var outputFormat string
	var sessionID string

	cmd := &cobra.Command{
		Use:   "actions",
		Short: "List stored agent actions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := loadAgentStore(storePath)
			if err != nil {
				return err
			}

			var actions []agent.Action
			if sessionID != "" {
				actions = store.ListSessionActions(sessionID)
			} else {
				actions = store.ListActions()
			}

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), actions, outputFormat)
			}

			if len(actions) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No agent actions found.")
				return nil
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", "Actions")
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", "-------")
			for _, action := range actions {
				fmt.Fprintf(cmd.OutOrStdout(), "%s | session=%s | type=%s | target=%s | decision=%s\n",
					action.ID,
					action.SessionID,
					action.Type,
					action.Target,
					action.Decision,
				)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&storePath, "store-path", "", "path to the agent session store JSON file")
	cmd.Flags().StringVar(&sessionID, "session-id", "", "optional session identifier to filter actions")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "output format: text, json, yaml")
	return cmd
}

func newAgentDeleteCommand() *cobra.Command {
	var storePath string
	var sessionID string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a stored agent session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sessionID == "" {
				return fmt.Errorf("--session-id is required")
			}

			store, err := loadAgentStore(storePath)
			if err != nil {
				return err
			}
			if err := store.DeleteSession(sessionID); err != nil {
				return err
			}
			if err := store.Save(); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Deleted session %s\n", sessionID)
			return nil
		},
	}

	cmd.Flags().StringVar(&storePath, "store-path", "", "path to the agent session store JSON file")
	cmd.Flags().StringVar(&sessionID, "session-id", "", "session identifier (required)")
	_ = cmd.MarkFlagRequired("session-id")
	return cmd
}

func loadAgentStore(storePath string) (*agent.Store, error) {
	path := storePath
	if path == "" {
		path = defaultAgentStorePath()
	}
	store := agent.NewStore(path)
	if err := store.Load(); err != nil {
		return nil, err
	}
	return store, nil
}

func defaultAgentStorePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "naeos-agent-store.json")
	}
	return filepath.Join(home, ".config", "naeos", "agent-store.json")
}
