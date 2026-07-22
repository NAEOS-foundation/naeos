package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/internal/workflow"
)

const workflowStoreDir = ".naeos/workflows"

func workflowManager() *workflow.Manager {
	home, _ := os.UserHomeDir()
	storePath := filepath.Join(home, workflowStoreDir)
	return workflow.NewManagerWithPath(storePath)
}

func newWorkflowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Workflow and approval management",
		Long: `Create, execute, and manage workflows and approval processes.

Example:
  naeos workflow list
  naeos workflow create --name deploy-prod --steps build,test,deploy
  naeos workflow execute --name deploy-prod
  naeos workflow approve --id req-123 --approver admin --comment "LGTM"
  naeos workflow requests`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newWorkflowListCommand())
	cmd.AddCommand(newWorkflowCreateCommand())
	cmd.AddCommand(newWorkflowExecuteCommand())
	cmd.AddCommand(newWorkflowApproveCommand())
	cmd.AddCommand(newWorkflowRejectCommand())
	cmd.AddCommand(newWorkflowRequestsCommand())

	return cmd
}

func newWorkflowListCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all workflows",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := workflowManager()

			names := mgr.List()

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), names, outputFormat)
			}

			if len(names) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No workflows defined.")
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-20s\n", "WORKFLOW")
			fmt.Fprintf(out, "%-20s\n", "--------")
			for _, name := range names {
				fmt.Fprintf(out, "%-20s\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	return cmd
}

func newWorkflowCreateCommand() *cobra.Command {
	var name string
	var steps []string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workflow",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := workflowManager()

			var workflowSteps []*workflow.WorkflowStep
			for i, step := range steps {
				stepName := step
				workflowSteps = append(workflowSteps, &workflow.WorkflowStep{
					Name:     stepName,
					Action:   func(ctx *workflow.WorkflowContext) error { return nil },
					Timeout:  300,
					Required: true,
				})
				_ = i
			}

			w := workflow.NewWorkflow(name, workflowSteps)
			if err := mgr.Register(name, w); err != nil {
				return fmt.Errorf("register workflow: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created workflow '%s' with %d steps: %s\n",
				name, len(steps), strings.Join(steps, ", "))
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "workflow name (required)")
	cmd.Flags().StringArrayVar(&steps, "steps", nil, "workflow steps (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("steps")
	return cmd
}

func newWorkflowExecuteCommand() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute a workflow",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			mgr := workflowManager()

			if err := mgr.Execute(name); err != nil {
				return fmt.Errorf("execution failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Workflow '%s' executed successfully.\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "workflow name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newWorkflowApproveCommand() *cobra.Command {
	var id, approver, comment string

	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve a pending request",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			aw := workflow.NewApprovalWorkflow()

			if err := aw.Approve(id, approver, comment); err != nil {
				return fmt.Errorf("approval failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Request '%s' approved by %s.\n", id, approver)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "request ID (required)")
	cmd.Flags().StringVar(&approver, "approver", "", "approver name (required)")
	cmd.Flags().StringVar(&comment, "comment", "", "approval comment")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("approver")
	return cmd
}

func newWorkflowRejectCommand() *cobra.Command {
	var id, approver, comment string

	cmd := &cobra.Command{
		Use:   "reject",
		Short: "Reject a pending request",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			aw := workflow.NewApprovalWorkflow()

			if err := aw.Reject(id, approver, comment); err != nil {
				return fmt.Errorf("rejection failed: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Request '%s' rejected by %s.\n", id, approver)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "request ID (required)")
	cmd.Flags().StringVar(&approver, "approver", "", "approver name (required)")
	cmd.Flags().StringVar(&comment, "comment", "", "rejection comment")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("approver")
	return cmd
}

func newWorkflowRequestsCommand() *cobra.Command {
	var status string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "requests",
		Short: "List approval requests",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			aw := workflow.NewApprovalWorkflow()

			requests := aw.ListByStatus(status)

			if outputFormat != "" && outputFormat != "text" {
				return FormatOutput(cmd.OutOrStdout(), requests, outputFormat)
			}

			if len(requests) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No %s requests.\n", status)
				return nil
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%-10s %-20s %-15s %s\n", "ID", "WORKFLOW", "REQUESTER", "STATUS")
			fmt.Fprintf(out, "%-10s %-20s %-15s %s\n", "--", "--------", "---------", "------")
			for _, req := range requests {
				fmt.Fprintf(out, "%-10s %-20s %-15s %s\n",
					req.ID, req.Workflow, req.Requester, req.Status)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "pending", "filter by status (pending, approved, rejected)")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "", "output format: text, json, yaml")
	return cmd
}
