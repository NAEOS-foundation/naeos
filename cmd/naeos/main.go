package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/NAEOS-foundation/naeos/pkg/pipeline"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: naeos <init|run|validate>")
	}

	subcommand := args[0]
	switch subcommand {
	case "init":
		return runInit(args[1:])
	case "run":
		return runPipeline(args[1:])
	case "validate":
		return runValidate(args[1:])
	default:
		return fmt.Errorf("unknown command %q", subcommand)
	}
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	output := fs.String("output", "config.example.yaml", "path for the generated config file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	content := strings.Join([]string{
		"pipeline:",
		"  name: naeos-dev",
		"  mode: development",
		"  verbose: true",
		"  output_dir: ./out",
	}, "\n") + "\n"

	if err := os.WriteFile(*output, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("created %s\n", *output)
	return nil
}

func runPipeline(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to JSON or YAML config file")
	input := fs.String("input", "", "specification input to process")
	outputFormat := fs.String("output", "text", "output format: text, json, or yaml")
	outputFile := fs.String("output-file", "", "optional file path to write the formatted output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return fmt.Errorf("missing required --config")
	}
	if *input == "" {
		return fmt.Errorf("missing required --input")
	}

	cfg, err := pipeline.ConfigFromFile(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	p := pipeline.New(cfg)
	result, err := p.Run(*input)
	if err != nil {
		return fmt.Errorf("pipeline run failed: %w", err)
	}

	payload := map[string]any{
		"pipeline":   cfg.Name,
		"mode":       cfg.Mode,
		"verbose":    cfg.Verbose,
		"output_dir": cfg.OutputDir,
		"artifacts":  len(result.Artifacts),
		"tasks":      len(result.Tasks),
	}

	var rendered []byte
	switch strings.ToLower(*outputFormat) {
	case "json":
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return fmt.Errorf("encode json output: %w", err)
		}
		rendered = append(data, '\n')
	case "yaml":
		data, err := yaml.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode yaml output: %w", err)
		}
		rendered = data
	default:
		rendered = []byte(fmt.Sprintf("pipeline=%s mode=%s verbose=%t output_dir=%s\nartifacts=%d tasks=%d\n", result.NEIR.Project, cfg.Mode, cfg.Verbose, cfg.OutputDir, len(result.Artifacts), len(result.Tasks)))
	}

	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, rendered, 0o644); err != nil {
			return fmt.Errorf("write output file: %w", err)
		}
	} else {
		if _, err := os.Stdout.Write(rendered); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
	}
	return nil
}

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	configPath := fs.String("config", "", "path to JSON or YAML config file")
	input := fs.String("input", "", "specification input to process")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *configPath == "" {
		return fmt.Errorf("missing required --config")
	}
	if *input == "" {
		return fmt.Errorf("missing required --input")
	}

	cfg, err := pipeline.ConfigFromFile(*configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	_ = cfg
	fmt.Printf("config loaded successfully from %s\n", *configPath)
	fmt.Printf("input received: %s\n", *input)
	return nil
}
