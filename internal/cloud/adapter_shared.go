package cloud

import (
	"log/slog"
	"time"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type deployConfig struct {
	Provider       CloudProvider
	TempDirPrefix  string
	ResourceIDFunc func(name, resType string) string
}

func deployAdapter(a CloudAdapter, config *DeployConfig, dc deployConfig) (*DeployResult, error) {
	if err := a.Validate(config); err != nil {
		return nil, err
	}

	planResult, err := a.Plan(config)
	if err != nil {
		return nil, err
	}

	tf, err := a.ExportTerraform(config)
	if err != nil {
		return nil, err
	}

	pool := GetDefaultPool()
	tr, pooled := pool.Get(config.Project, dc.Provider)
	if pooled {
		if err := tr.writeHCL(tf); err != nil {
			return nil, err
		}
		if err := tr.Apply(); err != nil {
			return nil, err
		}
	} else {
		workDir, err := TempWorkDir(dc.TempDirPrefix)
		if err != nil {
			return nil, err
		}
		tr = NewTerraformRunner(workDir)
		if r, ok := a.(interface{ SetRunner(CommandRunner) }); ok {
			_ = r
		}
		if runner := getRunner(a); runner != nil {
			tr.Runner = runner
		}
		if err := tr.Deploy(tf); err != nil {
			return nil, err
		}
		pool.Put(config.Project, dc.Provider, tr, true)
	}

	deployed := make([]DeployedResource, 0, len(planResult.Resources))
	for _, res := range planResult.Resources {
		deployed = append(deployed, DeployedResource{
			Name: res.Name,
			Type: res.Type,
			ID:   dc.ResourceIDFunc(res.Name, res.Type),
		})
	}

	result := &DeployResult{
		Provider:  dc.Provider,
		Resources: deployed,
		Terraform: tf,
		Status:    "deployed",
		Timestamp: time.Now(),
	}

	sm := NewStateManager()
	_ = sm.Save(&DeploymentRecord{
		Project:      config.Project,
		Provider:     dc.Provider,
		Environment:  config.Environment,
		Region:       config.Region,
		Resources:    deployed,
		TerraformDir: tr.WorkDir,
		Timestamp:    result.Timestamp,
		Status:       "deployed",
	})

	return result, nil
}

func destroyAdapter(a CloudAdapter, config *DeployConfig, provider CloudProvider, tempDirPrefix string) error {
	pool := GetDefaultPool()
	if tr, pooled := pool.Get(config.Project, provider); pooled {
		if err := tr.ApplyDestroy(); err == nil {
			pool.Remove(config.Project, provider)
			sm := NewStateManager()
			if err := sm.Delete(config.Project, provider); err != nil {
				slog.Warn("failed to delete state after pool destroy", "provider", provider, "project", config.Project, "error", err)
			}
			return nil
		}
	}

	sm := NewStateManager()
	record, err := sm.Load(config.Project, provider)
	if err == nil && record.TerraformDir != "" {
		tr := NewTerraformRunner(record.TerraformDir)
		if runner := getRunner(a); runner != nil {
			tr.Runner = runner
		}
		if derr := tr.DestroyAll(); derr == nil {
			if err := sm.Delete(config.Project, provider); err != nil {
				slog.Warn("failed to delete state after destroy", "provider", provider, "project", config.Project, "error", err)
			}
			return nil
		}
	}

	planResult, err := a.Plan(config)
	if err != nil {
		return err
	}
	if len(planResult.Resources) == 0 {
		return naeoserr.New(naeoserr.ErrValidation, "no resources to destroy")
	}

	tf, err := a.ExportTerraform(config)
	if err != nil {
		return err
	}

	workDir, werr := TempWorkDir(tempDirPrefix)
	if werr != nil {
		return werr
	}

	tr := NewTerraformRunner(workDir)
	if runner := getRunner(a); runner != nil {
		tr.Runner = runner
	}
	if err := tr.writeHCL(tf); err != nil {
		return err
	}
	if err := tr.DestroyAll(); err != nil {
		return err
	}

	if err := sm.Delete(config.Project, provider); err != nil {
		slog.Warn("failed to delete state after terraform destroy", "provider", provider, "project", config.Project, "error", err)
	}
	return nil
}

func getRunner(a CloudAdapter) CommandRunner {
	switch v := a.(type) {
	case *AWSAdapter:
		return v.Runner
	case *GCPAdapter:
		return v.Runner
	case *AzureAdapter:
		return v.Runner
	default:
		return nil
	}
}
