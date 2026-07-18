package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type DeployTarget struct {
	Name    string
	Command string
	Args    []string
}

func newDeployCommand() *cobra.Command {
	var configPath string
	var target string
	var environment string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the pipeline output to a target environment",
		Long: `Deploy generated artifacts to a target environment using configured deployment tools.

Supported targets:
  docker    — Build and push Docker images
  k8s       — Apply Kubernetes manifests
  compose   — Docker Compose up
  ssh       — Remote deployment via SSH
  rsync     — File sync via rsync
  local     — Local directory copy

Example:
  naeos deploy --target docker
  naeos deploy --target k8s --env staging
  naeos deploy --target compose --dry-run
  naeos deploy --target rsync --env production`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(cmd, configPath, target, environment, dryRun)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to config file")
	cmd.Flags().StringVarP(&target, "target", "t", "local", "deployment target: docker, k8s, compose, ssh, rsync, local")
	cmd.Flags().StringVarP(&environment, "env", "e", "development", "target environment")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview deployment without executing")
	return cmd
}

func runDeploy(cmd *cobra.Command, configPath, target, environment string, dryRun bool) error {
	out := cmd.OutOrStdout()

	cfg, _, err := loadPipelineConfig(configPath, cliVerbose, nil, cliDryRun)
	if err != nil {
		return err
	}

	outputDir := cfg.OutputDir
	if outputDir == "" {
		outputDir = "./output"
	}

	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return fmt.Errorf("output directory %q does not exist. Run 'naeos run' first", outputDir)
	}

	fmt.Fprintf(out, "Deploying %s to %s (env: %s)\n", cfg.Name, target, environment)
	fmt.Fprintf(out, "Output directory: %s\n", outputDir)

	var deployCmd *DeployTarget
	switch target {
	case "docker":
		deployCmd = &DeployTarget{Name: "Docker", Command: "docker", Args: []string{"build", "-t", cfg.Name + ":" + environment, "."}}
	case "k8s":
		manifests := filepath.Join(outputDir, "k8s")
		deployCmd = &DeployTarget{Name: "Kubernetes", Command: "kubectl", Args: []string{"apply", "-f", manifests}}
	case "compose":
		deployCmd = &DeployTarget{Name: "Docker Compose", Command: "docker-compose", Args: []string{"up", "-d"}}
	case "ssh":
		deployCmd = &DeployTarget{Name: "SSH", Command: "rsync", Args: []string{"-avz", outputDir + "/", environment + ":/app/"}}
	case "rsync":
		deployCmd = &DeployTarget{Name: "rsync", Command: "rsync", Args: []string{"-avz", "--delete", outputDir + "/", environment + ":/deploy/"}}
	case "local":
		fmt.Fprintf(out, "Local deployment: copying %s to %s-deploy/\n", outputDir, cfg.Name)
		deployDir := cfg.Name + "-deploy"
		if dryRun {
			fmt.Fprintf(out, "[dry-run] Would copy %s → %s/\n", outputDir, deployDir)
			return nil
		}
		return cpDir(outputDir, deployDir)
	default:
		return fmt.Errorf("unknown target %q. Supported: docker, k8s, compose, ssh, rsync, local", target)
	}

	if dryRun {
		fmt.Fprintf(out, "[dry-run] Would execute: %s %s\n", deployCmd.Command, strings.Join(deployCmd.Args, " "))
		return nil
	}

	fmt.Fprintf(out, "Executing: %s %s\n", deployCmd.Command, strings.Join(deployCmd.Args, " "))
	c := exec.Command(deployCmd.Command, deployCmd.Args...)
	c.Stdout = out
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("deploy failed: %w", err)
	}

	fmt.Fprintf(out, "✓ Deployment to %s complete\n", target)
	return nil
}

func cpDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		perm := info.Mode().Perm()
		if perm&0o111 != 0 {
			perm |= 0o111
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
		if err != nil {
			return err
		}
		if _, err := file.Write(data); err != nil {
			file.Close()
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
		return os.Chmod(target, perm)
	})
}
