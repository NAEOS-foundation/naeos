package hcl

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Spec struct {
	Project  Project            `json:"project"`
	Services map[string]Service `json:"services"`
	Infra    Infra              `json:"infra"`
}

type Project struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

type Service struct {
	Image string `json:"image,omitempty"`
	Port  int    `json:"port,omitempty"`
	Type  string `json:"type"`
}

type Infra struct {
	Engine string `json:"engine,omitempty"`
}

func ParseFile(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return Parse(data, path)
}

var blockRe = regexp.MustCompile(`^(\w+)\s+"([^"]*)"\s*\{`)
var kvRe = regexp.MustCompile(`^\s*(\w+)\s*=\s*(.+)$`)

func Parse(data []byte, filename string) (*Spec, error) {
	spec := &Spec{Services: make(map[string]Service)}
	lines := strings.Split(string(data), "\n")

	var currentBlock string
	var currentLabel string

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		if line == "}" {
			currentBlock = ""
			currentLabel = ""
			continue
		}

		if m := blockRe.FindStringSubmatch(line); m != nil {
			currentBlock = m[1]
			currentLabel = m[2]
			if currentBlock == "project" {
				spec.Project.Name = currentLabel
			}
			continue
		}

		if m := kvRe.FindStringSubmatch(line); m != nil && currentBlock != "" {
			key := m[1]
			val := strings.TrimSpace(m[2])
			val = strings.Trim(val, `"`)

			switch currentBlock {
			case "project":
				switch key {
				case "version":
					spec.Project.Version = val
				case "description":
					spec.Project.Description = val
				}
			case "service":
				svc := spec.Services[currentLabel]
				switch key {
				case "image":
					svc.Image = val
				case "port":
					p, _ := strconv.Atoi(val)
					svc.Port = p
				case "type":
					svc.Type = val
				}
				spec.Services[currentLabel] = svc
			case "infra":
				switch key {
				case "engine":
					spec.Infra.Engine = val
				}
			}
			continue
		}

		_ = i
	}

	return spec, nil
}

func ToYAML(spec *Spec) ([]byte, error) {
	var out []byte
	out = append(out, []byte("project:\n")...)
	out = append(out, []byte(fmt.Sprintf("  name: %s\n", spec.Project.Name))...)
	if spec.Project.Version != "" {
		out = append(out, []byte(fmt.Sprintf("  version: %s\n", spec.Project.Version))...)
	}
	if spec.Project.Description != "" {
		out = append(out, []byte(fmt.Sprintf("  description: %s\n", spec.Project.Description))...)
	}

	if len(spec.Services) > 0 {
		out = append(out, []byte("services:\n")...)
		for name, svc := range spec.Services {
			out = append(out, []byte(fmt.Sprintf("  - name: %s\n", name))...)
			if svc.Image != "" {
				out = append(out, []byte(fmt.Sprintf("    image: %s\n", svc.Image))...)
			}
			if svc.Port != 0 {
				out = append(out, []byte(fmt.Sprintf("    port: %d\n", svc.Port))...)
			}
			if svc.Type != "" {
				out = append(out, []byte(fmt.Sprintf("    type: %s\n", svc.Type))...)
			}
		}
	}

	if spec.Infra.Engine != "" {
		out = append(out, []byte("infra:\n")...)
		out = append(out, []byte(fmt.Sprintf("  engine: %s\n", spec.Infra.Engine))...)
	}

	return out, nil
}
