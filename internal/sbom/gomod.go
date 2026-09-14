// Copyright 2024-2026 NAEOS Foundation
// SPDX-License-Identifier: Apache-2.0

package sbom

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

// goModuleLicense is a curated, audit-verified SPDX license map for the Go
// modules pinned by the NAEOS go.mod. It is maintained manually and does not
// replace an automated resolver; modules not listed here carry no license
// field in the generated BOM.
var goModuleLicense = map[string]string{
	"github.com/DATA-DOG/go-sqlmock":        "MIT",
	"github.com/bwmarrin/discordgo":         "BSD-3-Clause",
	"github.com/charmbracelet/bubbletea":    "MIT",
	"github.com/charmbracelet/lipgloss":     "MIT",
	"github.com/fsnotify/fsnotify":          "BSD-3-Clause",
	"github.com/go-sql-driver/mysql":        "MPL-2.0",
	"github.com/gorilla/websocket":          "BSD-2-Clause",
	"github.com/jackc/pgx/v5":               "MIT",
	"github.com/nats-io/nats.go":            "Apache-2.0",
	"github.com/rabbitmq/amqp091-go":        "BSD-2-Clause",
	"github.com/redis/go-redis/v9":          "BSD-2-Clause",
	"github.com/segmentio/kafka-go":         "MIT",
	"github.com/spf13/cobra":                "Apache-2.0",
	"github.com/tetratelabs/wazero":         "Apache-2.0",
	"golang.org/x/crypto":                   "BSD-3-Clause",
	"golang.org/x/mod":                      "BSD-3-Clause",
	"golang.org/x/sync":                     "BSD-3-Clause",
	"golang.org/x/text":                     "BSD-3-Clause",
	"gopkg.in/yaml.v3":                      "MIT AND Apache-2.0",
	"modernc.org/sqlite":                    "BSD-3-Clause",
	"filippo.io/edwards25519":               "BSD-3-Clause",
	"github.com/aymanbagabas/go-osc52/v2":   "MIT",
	"github.com/cespare/xxhash/v2":          "MIT",
	"github.com/charmbracelet/colorprofile": "MIT",
	"github.com/charmbracelet/x/ansi":       "MIT",
	"github.com/charmbracelet/x/cellbuf":    "MIT",
	"github.com/charmbracelet/x/term":       "MIT",
	"github.com/cpuguy83/go-md2man/v2":      "MIT",
	"github.com/dustin/go-humanize":         "MIT",
	"github.com/erikgeiser/coninput":        "MIT",
	"github.com/google/uuid":                "BSD-3-Clause",
	"github.com/inconshreveable/mousetrap":  "Apache-2.0",
	"github.com/jackc/pgpassfile":           "MIT",
	"github.com/jackc/pgservicefile":        "MIT",
	"github.com/jackc/puddle/v2":            "MIT",
	"github.com/klauspost/compress":         "Apache-2.0",
	"github.com/kr/text":                    "MIT",
	"github.com/lucasb-eyer/go-colorful":    "MIT",
	"github.com/mattn/go-isatty":            "MIT",
	"github.com/mattn/go-localereader":      "MIT",
	"github.com/mattn/go-runewidth":         "MIT",
	"github.com/muesli/ansi":                "MIT",
	"github.com/muesli/cancelreader":        "MIT",
	"github.com/muesli/termenv":             "MIT",
	"github.com/nats-io/nkeys":              "Apache-2.0",
	"github.com/nats-io/nuid":               "Apache-2.0",
	"github.com/ncruces/go-strftime":        "MIT",
	"github.com/pierrec/lz4/v4":             "BSD-3-Clause",
	"github.com/remyoudompheng/bigfft":      "BSD-3-Clause",
	"github.com/rivo/uniseg":                "MIT",
	"github.com/rogpeppe/go-internal":       "BSD-3-Clause",
	"github.com/russross/blackfriday/v2":    "BSD-2-Clause",
	"github.com/spf13/pflag":                "BSD-3-Clause",
	"github.com/xo/terminfo":                "MIT",
	"go.uber.org/atomic":                    "MIT",
	"go.yaml.in/yaml/v3":                    "MIT AND Apache-2.0",
	"golang.org/x/sys":                      "BSD-3-Clause",
	"modernc.org/libc":                      "BSD-3-Clause",
	"modernc.org/mathutil":                  "BSD-3-Clause",
	"modernc.org/memory":                    "BSD-3-Clause",
}

// moduleRef is a Go module pinned by go.mod.
type moduleRef struct {
	Path     string
	Version  string
	Indirect bool
}

// parseGoMod extracts the module requirements declared in a go.mod file.
func parseGoMod(data []byte) []moduleRef {
	var refs []moduleRef
	s := bufio.NewScanner(strings.NewReader(string(data)))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] == "module" || fields[0] == "go" ||
			fields[0] == "toolchain" || fields[0] == "retract" || fields[0] == "exclude" {
			continue
		}
		path, version := fields[0], fields[1]
		// Single-line form: require <path> <version>
		if path == "require" && len(fields) >= 3 {
			path, version = fields[1], fields[2]
		}
		if !strings.HasPrefix(version, "v") {
			continue
		}
		ref := moduleRef{Path: path, Version: version}
		if strings.Contains(line, "// indirect") {
			ref.Indirect = true
		}
		refs = append(refs, ref)
	}
	return refs
}

// parseGoSum returns the content hash for each module@version pinned in a
// go.sum file, excluding the pseudo-entries for module go.mod files.
func parseGoSum(data []byte) map[string]string {
	sums := make(map[string]string)
	s := bufio.NewScanner(strings.NewReader(string(data)))
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		path := fields[0]
		version := fields[1]
		if strings.HasSuffix(version, "/go.mod") {
			continue
		}
		hash := fields[2]
		if !strings.HasPrefix(hash, "h1:") {
			continue
		}
		key := path + "@" + version
		if _, ok := sums[key]; !ok {
			sums[key] = strings.TrimPrefix(hash, "h1:")
		}
	}
	return sums
}

// FromGoModules builds a dependency-level CycloneDX BOM from the go.mod and
// go.sum manifests in root. Direct and indirect modules are emitted as
// library components with Go package-urls, go.sum content hashes where
// available, and audit-verified SPDX licenses for curated modules.
func (g *Generator) FromGoModules(root string) (*BOM, error) {
	goModPath := filepath.Join(root, "go.mod")
	goSumPath := filepath.Join(root, "go.sum")

	modData, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "read %s", goModPath)
	}
	sumData, err := os.ReadFile(goSumPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "read %s", goSumPath)
	}

	refs := parseGoMod(modData)
	sums := parseGoSum(sumData)

	comps := make([]Component, 0, len(refs))
	for _, ref := range refs {
		comp := Component{
			Type:     Library,
			Group:    "golang",
			Name:     ref.Path,
			Version:  ref.Version,
			Purl:     Purl("golang", ref.Path, ref.Version),
			FileName: ref.Path + "@" + ref.Version,
			Path:     "go.mod",
			License:  goModuleLicense[ref.Path],
			Properties: []Property{
				{Name: "go.module", Value: "direct"},
			},
		}
		if ref.Indirect {
			comp.Properties[0].Value = "indirect"
		}
		if sum, ok := sums[ref.Path+"@"+ref.Version]; ok {
			comp.Hashes = []Hash{{Alg: "H1", Val: sum}}
		}
		comps = append(comps, comp)
	}

	bom, err := g.Generate(comps)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "generate module BOM")
	}
	if bom.Metadata.Component != nil {
		bom.Metadata.Component.Purl = Purl("pkg", g.cfg.Project, g.cfg.Version)
	}
	bom.Dependencies = []Dependency{{Ref: goModuleFileRef(root, g.cfg.Project)}}
	return bom, nil
}

func goModuleFileRef(root, project string) string {
	if project != "" {
		return project
	}
	return fmt.Sprintf("%s/go.mod", filepath.Base(root))
}

// GoModuleLicensed reports whether a SPDX license is curated for a module.
func GoModuleLicensed(path string) bool {
	_, ok := goModuleLicense[path]
	return ok
}
