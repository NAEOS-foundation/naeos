package renderers

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type Renderer interface {
	Render(tmpl string, data any) ([]byte, error)
	RenderNamed(name, tmpl string, data any) ([]byte, error)
}

type DefaultRenderer struct{}

func NewRenderer() Renderer {
	return DefaultRenderer{}
}

func (DefaultRenderer) Render(tmpl string, data any) ([]byte, error) {
	return RenderTemplate("default", tmpl, data)
}

func (DefaultRenderer) RenderNamed(name, tmpl string, data any) ([]byte, error) {
	return RenderTemplate(name, tmpl, data)
}

func RenderTemplate(name, tmpl string, data any) ([]byte, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "parse template %s", name)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "execute template %s", name)
	}

	return buf.Bytes(), nil
}

func RenderWithFuncs(name, tmpl string, data any, funcs template.FuncMap) ([]byte, error) {
	t, err := template.New(name).Funcs(funcs).Parse(tmpl)
	if err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "parse template %s", name)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "execute template %s", name)
	}

	return buf.Bytes(), nil
}

type TemplateData struct {
	Project    string
	Module     string
	Service    string
	Port       int
	Kind       string
	Version    string
	Package    string
	Language   string
	Attributes map[string]string
}

type CodeRenderer interface {
	RenderCode(lang string, tmpl string, data any) ([]byte, error)
	FormatCode(lang string, code []byte) ([]byte, error)
}

type GoRenderer struct{}

func NewGoRenderer() CodeRenderer {
	return GoRenderer{}
}

func (g GoRenderer) RenderCode(lang string, tmpl string, data any) ([]byte, error) {
	raw, err := RenderTemplate("go:"+lang, tmpl, data)
	if err != nil {
		return nil, err
	}
	return g.FormatCode(lang, raw)
}

func (g GoRenderer) FormatCode(lang string, code []byte) ([]byte, error) {
	formatted, err := format.Source(code)
	if err != nil {
		return code, naeoserr.Wrapf(err, naeoserr.ErrInternal, "go fmt failed")
	}
	return formatted, nil
}

type TypeScriptRenderer struct{}

func NewTypeScriptRenderer() CodeRenderer {
	return TypeScriptRenderer{}
}

func (ts TypeScriptRenderer) RenderCode(lang string, tmpl string, data any) ([]byte, error) {
	return RenderTemplate("ts:"+lang, tmpl, data)
}

func (ts TypeScriptRenderer) FormatCode(lang string, code []byte) ([]byte, error) {
	return code, nil
}

type PythonRenderer struct{}

func NewPythonRenderer() CodeRenderer {
	return PythonRenderer{}
}

func (p PythonRenderer) RenderCode(lang string, tmpl string, data any) ([]byte, error) {
	return RenderTemplate("py:"+lang, tmpl, data)
}

func (p PythonRenderer) FormatCode(lang string, code []byte) ([]byte, error) {
	return code, nil
}

func NewCodeRenderer(lang string) CodeRenderer {
	switch strings.ToLower(lang) {
	case "go", "golang":
		return NewGoRenderer()
	case "typescript", "ts", "tsx":
		return NewTypeScriptRenderer()
	case "python", "py":
		return NewPythonRenderer()
	default:
		return nil
	}
}

type FileSpec struct {
	Path    string
	Content []byte
	Perm    os.FileMode
}

type FileGenerator struct {
	files []FileSpec
}

func NewFileGenerator() *FileGenerator {
	return &FileGenerator{}
}

func (fg *FileGenerator) AddFile(path string, content []byte) {
	fg.files = append(fg.files, FileSpec{Path: path, Content: content, Perm: 0o644})
}

func (fg *FileGenerator) AddExecutable(path string, content []byte) {
	fg.files = append(fg.files, FileSpec{Path: path, Content: content, Perm: 0o755})
}

func (fg *FileGenerator) Files() []FileSpec {
	result := make([]FileSpec, len(fg.files))
	copy(result, fg.files)
	return result
}

func (fg *FileGenerator) WriteToDir(dir string) error {
	for _, f := range fg.files {
		fullPath := filepath.Join(dir, f.Path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrInternal, "create directory for %s", f.Path)
		}
		if err := os.WriteFile(fullPath, f.Content, f.Perm); err != nil {
			return naeoserr.Wrapf(err, naeoserr.ErrInternal, "write %s", f.Path)
		}
	}
	return nil
}

func (fg *FileGenerator) Count() int {
	return len(fg.files)
}

func (fg *FileGenerator) Paths() []string {
	var paths []string
	for _, f := range fg.files {
		paths = append(paths, f.Path)
	}
	return paths
}

func (fg *FileGenerator) Contains(path string) bool {
	for _, f := range fg.files {
		if f.Path == path {
			return true
		}
	}
	return false
}

type MultiRenderer struct {
	renderers map[string]CodeRenderer
}

func NewMultiRenderer() *MultiRenderer {
	return &MultiRenderer{
		renderers: make(map[string]CodeRenderer),
	}
}

func (mr *MultiRenderer) Register(lang string, renderer CodeRenderer) {
	mr.renderers[lang] = renderer
}

func (mr *MultiRenderer) Render(lang string, tmpl string, data any) ([]byte, error) {
	r, ok := mr.renderers[lang]
	if !ok {
		return nil, naeoserr.New(naeoserr.ErrInternal, fmt.Sprintf("no renderer registered for language %q", lang))
	}
	return r.RenderCode(lang, tmpl, data)
}

func (mr *MultiRenderer) Format(lang string, code []byte) ([]byte, error) {
	r, ok := mr.renderers[lang]
	if !ok {
		return nil, naeoserr.New(naeoserr.ErrInternal, fmt.Sprintf("no renderer registered for language %q", lang))
	}
	return r.FormatCode(lang, code)
}

func (mr *MultiRenderer) Languages() []string {
	var langs []string
	for lang := range mr.renderers {
		langs = append(langs, lang)
	}
	return langs
}

type RenderPipeline struct {
	steps []RenderStep
}

type RenderStep struct {
	Name string
	Fn   func(data any) ([]byte, error)
}

func NewRenderPipeline() *RenderPipeline {
	return &RenderPipeline{}
}

func (rp *RenderPipeline) AddStep(name string, fn func(data any) ([]byte, error)) {
	rp.steps = append(rp.steps, RenderStep{Name: name, Fn: fn})
}

func (rp *RenderPipeline) Execute(data any) ([][]byte, error) {
	var results [][]byte
	for _, step := range rp.steps {
		out, err := step.Fn(data)
		if err != nil {
			return nil, naeoserr.Wrapf(err, naeoserr.ErrInternal, "step %s failed", step.Name)
		}
		results = append(results, out)
	}
	return results, nil
}

func (rp *RenderPipeline) StepCount() int {
	return len(rp.steps)
}

type TemplateLibrary struct {
	templates map[string]string
}

func NewTemplateLibrary() *TemplateLibrary {
	return &TemplateLibrary{
		templates: make(map[string]string),
	}
}

func (tl *TemplateLibrary) Register(name, tmpl string) {
	tl.templates[name] = tmpl
}

func (tl *TemplateLibrary) Get(name string) (string, bool) {
	t, ok := tl.templates[name]
	return t, ok
}

func (tl *TemplateLibrary) Render(name string, data any) ([]byte, error) {
	tmpl, ok := tl.templates[name]
	if !ok {
		return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("template %q not found", name))
	}
	return RenderTemplate(name, tmpl, data)
}

func (tl *TemplateLibrary) Names() []string {
	var names []string
	for name := range tl.templates {
		names = append(names, name)
	}
	return names
}

func (tl *TemplateLibrary) Has(name string) bool {
	_, ok := tl.templates[name]
	return ok
}
