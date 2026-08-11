package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/spf13/cobra"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func newRunCommand() *cobra.Command {
	var configPath, input, inputFile, outputFormat, outputFile, profileOut, pprofAddr string
	var languages []string
	var dryRun, profiling bool

	var cacheDir string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute the NAEOS pipeline",
		Long: `Execute the full NAEOS pipeline: parse, normalize, resolve, build NEIR, generate artifacts.

Example:
  naeos run --config config.yaml --input spec.yaml
  naeos run --config config.yaml --input-file spec.yaml --output json
  naeos run --config config.yaml --input spec.yaml --language go --language typescript
  naeos run --config config.yaml --input spec.yaml --cache-dir .naeos/cache`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if pprofAddr != "" {
				mux := http.NewServeMux()
				mux.HandleFunc("/debug/pprof/", pprof.Index)
				mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
				mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
				mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
				mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
				go func() {
					fmt.Printf("pprof HTTP server listening on %s\n", pprofAddr)
					srv := &http.Server{
						Addr:              pprofAddr,
						Handler:           mux,
						ReadHeaderTimeout: 5 * time.Second,
						ReadTimeout:       30 * time.Second,
						WriteTimeout:      30 * time.Second,
						IdleTimeout:       60 * time.Second,
					}
					_ = srv.ListenAndServe()
				}()
			}

			inputValue, err := loadInput(input, inputFile)
			if err != nil {
				return err
			}

			cfg, err := loadPipelineConfig(configPath, cliVerbose, languages, cliDryRun || dryRun, cacheDir)
			if err != nil {
				return err
			}
			cfg.Profiling = profiling

			p, err := pipeline.New(*cfg)
			if err != nil {
				return fmt.Errorf("failed to construct pipeline: %w", err)
			}

			result, err := p.Run(inputValue)
			if err != nil {
				return fmt.Errorf("pipeline run failed: %w", err)
			}

			if profileOut != "" && p.ProfileEnabled() {
				if err := p.ProfileSave(profileOut); err != nil {
					return fmt.Errorf("save profile: %w", err)
				}
				fmt.Printf("profile saved to %s\n", profileOut)
			}

			payload := map[string]any{
				"pipeline":   cfg.Name,
				"mode":       cfg.Mode,
				"verbose":    cfg.Verbose,
				"output_dir": cfg.OutputDir,
				"artifacts":  len(result.Artifacts),
				"tasks":      len(result.Tasks),
			}

			if len(languages) > 0 {
				payload["languages"] = languages
			}
			if cfg.DryRun {
				payload["dry_run"] = true
			}

			rendered, err := renderOutput(payload, outputFormat, func() []byte {
				return []byte(fmt.Sprintf("pipeline=%s mode=%s verbose=%t output_dir=%s\nartifacts=%d tasks=%d\n", result.NEIR.Project, cfg.Mode, cfg.Verbose, cfg.OutputDir, len(result.Artifacts), len(result.Tasks)))
			})
			if err != nil {
				return err
			}

			return writeOrPrint(cmd, rendered, outputFile)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "path to JSON or YAML config file (auto-detected if omitted)")
	cmd.Flags().StringVar(&input, "input", "", "specification input to process")
	cmd.Flags().StringVar(&inputFile, "input-file", "", "path to a specification file")
	cmd.Flags().StringVar(&outputFormat, "output", "text", "output format: text, json, or yaml")
	cmd.Flags().StringVar(&outputFile, "output-file", "", "optional file path to write the formatted output")
	cmd.Flags().StringArrayVar(&languages, "language", nil, "target language for code generation (go, typescript, python, java, rust)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview artifacts without writing to disk")
	cmd.Flags().BoolVar(&profiling, "profile", false, "enable pipeline profiling")
	cmd.Flags().StringVar(&profileOut, "profile-out", "profile.json", "path to write profile JSON")
	cmd.Flags().StringVar(&pprofAddr, "pprof", "", "pprof HTTP server address (e.g. :6060)")
	cmd.Flags().StringVar(&cacheDir, "cache-dir", "", "enable pipeline caching using the given directory")
	return cmd
}
