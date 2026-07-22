package promptlib

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/NAEOS-foundation/naeos/internal/neir/model"
)

type errList []error

func (e errList) Error() string {
	if len(e) == 0 {
		return ""
	}
	msgs := make([]string, len(e))
	for i, err := range e {
		msgs[i] = err.Error()
	}
	return fmt.Sprintf("%d override load error(s): %s", len(e), strings.Join(msgs, "; "))
}

// Library is the central prompt template library.
// It manages LLM prompts and compiler templates, supporting built-in defaults
// and user-provided overrides.
type Library struct {
	mu           sync.RWMutex
	llmPrompts   map[string]*LLMPrompt
	compilerTpls map[string]*CompilerTemplate
	overridesDir string
}

// Option configures a Library.
type Option func(*options)

type options struct {
	overridesDir string
}

// WithOverridesDir sets the directory for user-provided prompt overrides.
func WithOverridesDir(dir string) Option {
	return func(o *options) {
		o.overridesDir = dir
	}
}

// New creates a new Library with the given options.
func New(opts ...Option) (*Library, error) {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	l := &Library{
		llmPrompts:   make(map[string]*LLMPrompt),
		compilerTpls: make(map[string]*CompilerTemplate),
		overridesDir: o.overridesDir,
	}

	if err := l.loadBuiltins(); err != nil {
		return nil, fmt.Errorf("load builtins: %w", err)
	}

	if o.overridesDir != "" {
		if err := l.loadOverrides(); err != nil {
			return nil, fmt.Errorf("load overrides: %w", err)
		}
	}

	return l, nil
}

// NewWithDefaults creates a Library with built-in prompts only.
func NewWithDefaults() *Library {
	l, _ := New()
	return l
}

func (l *Library) loadBuiltins() error {
	for name, data := range builtinLLMPrompts {
		p, err := ParseLLMPrompt([]byte(data))
		if err != nil {
			return fmt.Errorf("parse builtin LLM prompt %s: %w", name, err)
		}
		l.llmPrompts[name] = p
	}

	for name, data := range builtinCompilerTemplates {
		t, err := ParseCompilerTemplate([]byte(data))
		if err != nil {
			return fmt.Errorf("parse builtin compiler template %s: %w", name, err)
		}
		l.compilerTpls[name] = t
	}

	return nil
}

func (l *Library) loadOverrides() error {
	if l.overridesDir == "" {
		return nil
	}

	files, err := LoadPromptsFromDir(l.overridesDir)
	if err != nil {
		return err
	}

	var errs errList
	for path, data := range files {
		var meta struct {
			Kind string `yaml:"kind"`
			Name string `yaml:"name"`
		}
		if err := parseYAML(data, &meta); err != nil {
			errs = append(errs, fmt.Errorf("%s: parse meta: %w", path, err))
			continue
		}

		switch meta.Kind {
		case "llm":
			p, err := ParseLLMPrompt(data)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: parse LLM prompt: %w", path, err))
				continue
			}
			l.llmPrompts[p.Name] = p
		case "compiler":
			t, err := ParseCompilerTemplate(data)
			if err != nil {
				errs = append(errs, fmt.Errorf("%s: parse compiler template: %w", path, err))
				continue
			}
			l.compilerTpls[t.Name] = t
		default:
			errs = append(errs, fmt.Errorf("%s: unknown kind %q", path, meta.Kind))
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// GetLLMPrompt returns the named LLM prompt and true, or nil and false if not found.
func (l *Library) GetLLMPrompt(name string) (*LLMPrompt, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	p, ok := l.llmPrompts[name]
	return p, ok
}

// GetCompilerTemplate returns the named compiler template and true, or nil and false if not found.
func (l *Library) GetCompilerTemplate(name string) (*CompilerTemplate, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	t, ok := l.compilerTpls[name]
	return t, ok
}

// RenderLLM renders the named LLM prompt with the given variables.
func (l *Library) RenderLLM(name string, data map[string]any) (*RenderedLLM, error) {
	p, ok := l.GetLLMPrompt(name)
	if !ok {
		return nil, fmt.Errorf("LLM prompt %q not found", name)
	}
	return RenderLLM(p, data)
}

// RenderCompiler renders the named compiler template with NEIR data.
func (l *Library) RenderCompiler(name string, neir *model.NEIR) ([]RenderedFile, error) {
	t, ok := l.GetCompilerTemplate(name)
	if !ok {
		return nil, fmt.Errorf("compiler template %q not found", name)
	}
	return RenderCompiler(t, neir)
}

// ListLLMPrompts returns the names of all registered LLM prompts, sorted.
func (l *Library) ListLLMPrompts() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	names := make([]string, 0, len(l.llmPrompts))
	for name := range l.llmPrompts {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListCompilerTemplates returns the names of all registered compiler templates, sorted.
func (l *Library) ListCompilerTemplates() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	names := make([]string, 0, len(l.compilerTpls))
	for name := range l.compilerTpls {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegisterLLMPrompt registers a custom LLM prompt, overriding any existing one.
func (l *Library) RegisterLLMPrompt(name string, p *LLMPrompt) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if p != nil {
		p.Name = name
	}
	l.llmPrompts[name] = p
}

// RegisterCompilerTemplate registers a custom compiler template, overriding any existing one.
func (l *Library) RegisterCompilerTemplate(name string, t *CompilerTemplate) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if t != nil {
		t.Name = name
	}
	l.compilerTpls[name] = t
}

// ListAll returns metadata for all registered prompts and templates.
func (l *Library) ListAll() []PromptMeta {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []PromptMeta
	for _, p := range l.llmPrompts {
		result = append(result, PromptMeta{
			Name:        p.Name,
			Kind:        "llm",
			Version:     p.Version,
			Description: p.Description,
		})
	}
	for _, t := range l.compilerTpls {
		result = append(result, PromptMeta{
			Name:    t.Name,
			Kind:    "compiler",
			Version: t.Version,
			Target:  t.Target,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Kind != result[j].Kind {
			return result[i].Kind < result[j].Kind
		}
		return result[i].Name < result[j].Name
	})
	return result
}
