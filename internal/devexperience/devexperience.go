package devexperience

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// VS Code Extension

type VSCodeExtension struct {
	Name        string
	Version     string
	Description string
	Author      string
	Languages   []string
	Features    []string
}

func NewVSCodeExtension(name, version, description, author string, languages []string) *VSCodeExtension {
	return &VSCodeExtension{
		Name:        name,
		Version:     version,
		Description: description,
		Author:      author,
		Languages:   languages,
		Features:    []string{"syntax highlighting", "autocomplete", "linting"},
	}
}

func (e *VSCodeExtension) GenerateExtension(outputDir string) error {
	files := map[string]string{
		"package.json":                   e.GeneratePackageJSON(),
		"syntaxes/naeos.tmLanguage.json": e.GenerateSyntaxJSON(),
		"extension.js":                   e.GenerateExtensionJS(),
		".vscode/launch.json":            e.GenerateLaunchJSON(),
		"README.md":                      e.GenerateReadme(),
	}

	for relPath, content := range files {
		fullPath := filepath.Join(outputDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("create dir %s: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
			return fmt.Errorf("write %s: %w", relPath, err)
		}
	}

	return nil
}

func (e *VSCodeExtension) GeneratePackageJSON() string {
	languagesJSON := e.generateLanguagesJSON()
	return fmt.Sprintf(`{
  "name": "%[1]s",
  "displayName": "%[2]s",
  "description": "%[3]s",
  "version": "%[4]s",
  "publisher": "%[5]s",
  "license": "MIT",
  "engines": { "vscode": "^1.85.0" },
  "categories": ["Programming Languages", "Linters", "Language Packs"],
  "keywords": ["naeos", "neir", "specification", "yaml"],
  "activationEvents": [
    "onLanguage:naeos-yaml",
    "onCommand:naeos.compile",
    "onCommand:naeos.validate",
    "onCommand:naeos.dashboard"
  ],
  "main": "./extension.js",
  "contributes": {
    "languages": [%[6]s],
    "grammars": [{
      "language": "naeos-yaml",
      "scopeName": "source.naeos",
      "path": "./syntaxes/naeos.tmLanguage.json"
    }],
    "commands": [
      { "command": "naeos.compile",   "title": "NAEOS: Compile Project" },
      { "command": "naeos.validate",  "title": "NAEOS: Validate Spec" },
      { "command": "naeos.dashboard", "title": "NAEOS: Open Dashboard" },
      { "command": "naeos.lspStart",  "title": "NAEOS: Start Language Server" }
    ],
    "keybindings": [
      { "command": "naeos.compile",  "key": "ctrl+shift+b", "when": "editorLangId == naeos-yaml" },
      { "command": "naeos.validate", "key": "ctrl+shift+v", "when": "editorLangId == naeos-yaml" }
    ],
    "menus": {
      "editor/context": [
        { "command": "naeos.compile",  "group": "naeos" },
        { "command": "naeos.validate", "group": "naeos" }
      ]
    },
    "configuration": {
      "title": "NAEOS",
      "properties": {
        "naeos.lsp.path": {
          "type": "string",
          "default": "naeos",
          "description": "Path to the naeos CLI binary for the LSP server"
        },
        "naeos.compileOnSave": {
          "type": "boolean",
          "default": false,
          "description": "Compile the NAEOS project on file save"
        }
      }
    }
  }
}`, e.Name, e.displayName(), e.Description, e.Version, e.Author, languagesJSON)
}

func (e *VSCodeExtension) displayName() string {
	if e.Name == "" {
		return "Support"
	}
	return strings.ToUpper(e.Name[:1]) + e.Name[1:] + " Support"
}

func (e *VSCodeExtension) generateLanguagesJSON() string {
	return `{
  "id": "naeos-yaml",
  "aliases": ["NAEOS", "naeos-yaml"],
  "filenamePatterns": ["*.naeos.yaml", "*.naeos.yml"],
  "configuration": "./language-configuration.json"
}`
}

func (e *VSCodeExtension) GenerateSyntaxJSON() string {
	return `{
  "scopeName": "source.naeos",
  "fileTypes": ["naeos.yaml", "naeos.yml"],
  "name": "NAEOS Spec",
  "patterns": [
    { "include": "#comments" },
    { "include": "#top-level" },
    { "include": "#conditionals" },
    { "include": "#templates" }
  ],
  "repository": {
    "comments": {
      "patterns": [{
        "match": "#.*$",
        "name": "comment.line.number-sign.naeos"
      }]
    },
    "conditionals": {
      "patterns": [
        { "match": "\\$if\\{[^}]*\\}", "name": "keyword.control.conditional.naeos" },
        { "match": "\\$endif",          "name": "keyword.control.conditional.naeos" }
      ]
    },
    "templates": {
      "patterns": [
        { "match": "\\$\\{[^}]+\\}",     "name": "variable.other.naeos" },
        { "match": "\\$env\\{[^}]+\\}",  "name": "support.function.env.naeos" },
        { "match": "\\$ref\\{[^}]+\\}",  "name": "support.function.ref.naeos" },
        { "match": "\\$fn\\{[^}]+\\}",   "name": "support.function.fn.naeos" },
        { "match": "\\$import\\{[^}]+\\}","name": "support.function.import.naeos" },
        { "match": "\\$include\\{[^}]+\\}","name": "support.function.include.naeos" }
      ]
    },
    "top-level": {
      "patterns": [
        { "match": "^\\s*(project)(\\s*:)", "captures": { "1": { "name": "entity.name.tag.naeos" } } },
        { "match": "^\\s*(version|description|license)(\\s*:)", "captures": { "1": { "name": "support.type.property-name.naeos" } } },
        { "match": "^\\s*(modules|services|components|storage|api|endpoints)(\\s*:)", "captures": { "1": { "name": "entity.name.type.naeos" } } },
        { "match": "^\\s*(architecture|deployment|testing|generation|security|infrastructure|domain)(\\s*:)", "captures": { "1": { "name": "entity.name.namespace.naeos" } } },
        { "match": "^\\s*(name|path|kind|port|method|action|summary|pattern|strategy|provider|region|protocol)(\\s*:)", "captures": { "1": { "name": "variable.other.property.naeos" } } },
        { "match": "^\\s*(language|frameworks|principles|dependencies|tags|environments)(\\s*:)", "captures": { "1": { "name": "keyword.other.naeos" } } },
        { "match": "(http|grpc|worker|cli|job)\\b", "name": "constant.language.naeos" },
        { "match": "(layered|clean|hexagonal|microkernel|event-driven|cqrs|monolith|monolithic|microservices|serverless)\\b", "name": "constant.language.naeos" },
        { "match": "(rolling|blue-green|canary|recreate)\\b", "name": "constant.language.naeos" },
        { "match": "(go|typescript|python|java|rust)\\b", "name": "constant.language.naeos" },
        { "match": "(GET|POST|PUT|DELETE|PATCH)\\b", "name": "support.constant.http-method.naeos" },
        { "match": "(unit|integration|e2e|contract)\\b", "name": "constant.language.naeos" }
      ]
    }
  }
}`
}

func (e *VSCodeExtension) GenerateExtensionJS() string {
	// Use double-quoted Go string to avoid conflict with JS template literals (`)
	return "const vscode = require('vscode');\n" +
		"const { spawn } = require('child_process');\n" +
		"const path = require('path');\n" +
		"\n" +
		"function activate(context) {\n" +
		"    console.log('NAEOS extension activating...');\n" +
		"\n" +
		"    const compileCmd = vscode.commands.registerCommand('naeos.compile', async () => {\n" +
		"        const editor = vscode.window.activeTextEditor;\n" +
		"        if (!editor) return;\n" +
		"        const doc = editor.document;\n" +
		"        if (doc.languageId !== 'naeos-yaml') return;\n" +
		"        const terminal = vscode.window.createTerminal('NAEOS Compile');\n" +
		"        terminal.show();\n" +
		`        terminal.sendText('naeos build --input-file "' + doc.uri.fsPath + '"');\n` +
		"    });\n" +
		"\n" +
		"    const validateCmd = vscode.commands.registerCommand('naeos.validate', async () => {\n" +
		"        const editor = vscode.window.activeTextEditor;\n" +
		"        if (!editor) return;\n" +
		"        const doc = editor.document;\n" +
		"        if (doc.languageId !== 'naeos-yaml') return;\n" +
		"        const terminal = vscode.window.createTerminal('NAEOS Validate');\n" +
		"        terminal.show();\n" +
		`        terminal.sendText('naeos validate --input-file "' + doc.uri.fsPath + '"');\n` +
		"    });\n" +
		"\n" +
		"    const dashboardCmd = vscode.commands.registerCommand('naeos.dashboard', async () => {\n" +
		"        const terminal = vscode.window.createTerminal('NAEOS Dashboard');\n" +
		"        terminal.show();\n" +
		"        terminal.sendText('naeos dashboard');\n" +
		"    });\n" +
		"\n" +
		"    const lspStartCmd = vscode.commands.registerCommand('naeos.lspStart', () => {\n" +
		"        startLSPServer();\n" +
		"    });\n" +
		"\n" +
		"    const configListener = vscode.workspace.onDidChangeConfiguration(e => {\n" +
		"        if (e.affectsConfiguration('naeos')) {\n" +
		"            restartLSPServer();\n" +
		"        }\n" +
		"    });\n" +
		"\n" +
		"    startLSPServer();\n" +
		"\n" +
		"    context.subscriptions.push(compileCmd, validateCmd, dashboardCmd, lspStartCmd, configListener);\n" +
		"}\n" +
		"\n" +
		"let lspClient = null;\n" +
		"\n" +
		"function startLSPServer() {\n" +
		"    if (lspClient) return;\n" +
		"    const config = vscode.workspace.getConfiguration('naeos');\n" +
		"    const lspPath = config.get('lsp.path', 'naeos');\n" +
		"\n" +
		"    try {\n" +
		"        const workspaceFolders = vscode.workspace.workspaceFolders;\n" +
		"        const cwd = workspaceFolders ? workspaceFolders[0].uri.fsPath : undefined;\n" +
		"\n" +
		"        lspClient = spawn(lspPath, ['lsp'], { cwd, stdio: ['pipe', 'pipe', 'pipe'] });\n" +
		"\n" +
		"        lspClient.stdout.on('data', data => {\n" +
		"            console.log('[NAEOS LSP]', data.toString());\n" +
		"        });\n" +
		"\n" +
		"        lspClient.stderr.on('data', data => {\n" +
		"            console.log('[NAEOS LSP stderr]', data.toString());\n" +
		"        });\n" +
		"\n" +
		"        lspClient.on('error', err => {\n" +
		"            console.error('[NAEOS LSP] Failed to start:', err.message);\n" +
		`            vscode.window.showErrorMessage('NAEOS LSP: ' + err.message);\n` +
		"            lspClient = null;\n" +
		"        });\n" +
		"\n" +
		"        lspClient.on('exit', code => {\n" +
		"            console.log('[NAEOS LSP] exited with code', code);\n" +
		"            lspClient = null;\n" +
		"        });\n" +
		"\n" +
		"        vscode.window.showInformationMessage('NAEOS Language Server started');\n" +
		"    } catch (err) {\n" +
		"        console.error('[NAEOS LSP] Error:', err);\n" +
		`        vscode.window.showErrorMessage('NAEOS LSP error: ' + err.message);\n` +
		"    }\n" +
		"}\n" +
		"\n" +
		"function restartLSPServer() {\n" +
		"    if (lspClient) {\n" +
		"        lspClient.kill();\n" +
		"        lspClient = null;\n" +
		"    }\n" +
		"    startLSPServer();\n" +
		"}\n" +
		"\n" +
		"function deactivate() {\n" +
		"    if (lspClient) {\n" +
		"        lspClient.kill();\n" +
		"        lspClient = null;\n" +
		"    }\n" +
		"}\n" +
		"\n" +
		"module.exports = { activate, deactivate };\n"
}

func (e *VSCodeExtension) GenerateLaunchJSON() string {
	return `{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Run Extension",
      "type": "extensionHost",
      "request": "launch",
      "args": ["--extensionDevelopmentPath=\${workspaceFolder}"],
      "outFiles": ["\${workspaceFolder}/extension.js"]
    }
  ]
}`
}

func (e *VSCodeExtension) GenerateReadme() string {
	name := e.Name
	return "# " + e.displayName() + " Support for VS Code\n\n" +
		name + " extension for VS Code providing syntax highlighting, LSP integration,\n" +
		"and commands for the NAEOS declarative engineering platform.\n\n" +
		"## Features\n\n" +
		"- **Syntax highlighting** for `.naeos.yaml` and `.naeos.yml` files\n" +
		"- **LSP integration** \u2014 real-time diagnostics, autocomplete, hover, go-to-definition\n" +
		"- **Commands**: Compile, Validate, Open Dashboard\n" +
		"- **Keybindings**: `Ctrl+Shift+B` to compile, `Ctrl+Shift+V` to validate\n\n" +
		"## Requirements\n\n" +
		"- `naeos` CLI in $PATH (for LSP server and commands)\n\n" +
		"## Extension Settings\n\n" +
		"- `naeos.lsp.path`: Path to the `naeos` binary (default: `naeos`)\n" +
		"- `naeos.compileOnSave`: Auto-compile on file save (default: false)\n\n" +
		"## Developing\n\n" +
		"```bash\n" +
		"npm install -g vsce\n" +
		"vsce package\n" +
		"code --install-extension naeos-*.vsix\n" +
		"```\n"
}

// CLI Completion

type CompletionEngine struct {
	commands []string
	options  map[string][]string
}

func NewCompletionEngine() *CompletionEngine {
	e := &CompletionEngine{
		commands: []string{"init", "compile", "validate", "watch", "api", "ws", "graphql", "monitor", "auth", "db", "search", "workflow", "gateway", "cloud", "cicd", "pluginsdk"},
		options: map[string][]string{
			"init":      {"--name", "--type", "--language", "--framework"},
			"compile":   {"--input", "--output", "--language"},
			"validate":  {"--input", "--strict"},
			"watch":     {"--input", "--debounce"},
			"api":       {"--port", "--auth", "--secret"},
			"ws":        {"--port"},
			"graphql":   {"--port"},
			"monitor":   {"--port"},
			"auth":      {"login", "logout", "whoami"},
			"db":        {"connect", "disconnect", "migrate"},
			"search":    {"index", "query", "delete"},
			"workflow":  {"create", "list", "approve"},
			"gateway":   {"start", "stop"},
			"cloud":     {"deploy", "plan", "export"},
			"cicd":      {"generate", "list"},
			"pluginsdk": {"list", "info"},
		},
	}
	return e
}

func (e *CompletionEngine) Complete(input string) []string {
	parts := strings.Fields(input)

	if len(parts) == 0 {
		return e.commands
	}

	if len(parts) == 1 {
		var matches []string
		for _, cmd := range e.commands {
			if strings.HasPrefix(cmd, parts[0]) {
				matches = append(matches, cmd)
			}
		}
		return matches
	}

	if len(parts) == 2 {
		cmd := parts[0]
		if opts, ok := e.options[cmd]; ok {
			prefix := parts[1]
			var matches []string
			for _, opt := range opts {
				if strings.HasPrefix(opt, prefix) {
					matches = append(matches, opt)
				}
			}
			return matches
		}
	}

	return nil
}

func (e *CompletionEngine) GenerateBashCompletion() string {
	return `#!/bin/bash
_naeos_completions() {
    local cur prev commands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    commands="init compile validate watch api ws graphql monitor auth db search workflow gateway cloud cicd pluginsdk"
    
    if [ ${COMP_CWORD} -eq 1 ]; then
        COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
    fi
}
complete -F _naeos_completions naeos`
}

func (e *CompletionEngine) GenerateZshCompletion() string {
	return `#compdef naeos
_naeos() {
    _arguments \
        '1:command:(init compile validate watch api ws graphql monitor auth db search workflow gateway cloud cicd pluginsdk)' \
        '*::arg:->args'
}
_naeos "$@"`
}

func (e *CompletionEngine) GeneratePowerShellCompletion() string {
	return `Register-ArgumentCompleter -Native -CommandName naeos -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $commands = @('init', 'compile', 'validate', 'watch', 'api', 'ws', 'graphql', 'monitor', 'auth', 'db', 'search', 'workflow', 'gateway', 'cloud', 'cicd', 'pluginsdk')
    $commands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    }
}`
}

// Snippets

type SnippetManager struct {
	snippets map[string]string
}

func NewSnippetManager() *SnippetManager {
	sm := &SnippetManager{
		snippets: make(map[string]string),
	}

	sm.snippets["neir-spec"] = `project: my-project
version: 0.1.0
description: A new NAEOS project

modules:
  - name: core
    path: ./core

services:
  - name: api
    kind: http
    port: 8080

architecture:
  pattern: clean
  principles:
    - DI
    - SRP

deployment:
  strategy: rolling

generation:
  languages:
    - go
`
	sm.snippets["service"] = `  - name: service-name
    kind: http
    port: 8080
    description: Service description
    endpoints:
      - method: GET
        path: /resource
        action: listResources
`
	sm.snippets["module"] = `  - name: module-name
    path: ./internal/module-name
    description: Module description
    dependencies:
      - core
`

	return sm
}

func (sm *SnippetManager) Get(name string) (string, bool) {
	snippet, ok := sm.snippets[name]
	return snippet, ok
}

func (sm *SnippetManager) List() []string {
	names := make([]string, 0, len(sm.snippets))
	for name := range sm.snippets {
		names = append(names, name)
	}
	return names
}

func (sm *SnippetManager) Add(name, snippet string) {
	sm.snippets[name] = snippet
}

// Dev Experience Stack

type Stack struct {
	Extension *VSCodeExtension
	Engine    *CompletionEngine
	Snippets  *SnippetManager
}

func NewStack() *Stack {
	return &Stack{
		Extension: NewVSCodeExtension("naeos", "1.0.0", "NAEOS project support", "NAEOS", []string{"yaml", "json"}),
		Engine:    NewCompletionEngine(),
		Snippets:  NewSnippetManager(),
	}
}
