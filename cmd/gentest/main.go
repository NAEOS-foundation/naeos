package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

type FuncInfo struct {
	Name     string
	Params   []string
	Results  []string
	IsMethod bool
	RecvType string
}

type TypeInfo struct {
	Name   string
	Fields []string
}

type PackageInfo struct {
	Name    string
	Path    string
	Imports []string
	Funcs   []FuncInfo
	Types   []TypeInfo
}

var testTmpl = template.Must(template.New("test").Funcs(template.FuncMap{
	"isExported": func(s string) bool {
		if s == "" {
			return false
		}
		return s[0] >= 'A' && s[0] <= 'Z'
	},
	"hasPrefix": strings.HasPrefix,
}).Parse(`package {{.Name}}

import (
	"testing"
{{- range .Imports }}
	{{.}}
{{- end }}
)

{{- range .Funcs }}
{{- if isExported .Name }}
func Test{{.Name}}(t *testing.T) {
	t.Parallel()
	// TODO: implement test
}
{{- end }}
{{- end }}
`))

func main() {
	pkgPath := flag.String("pkg", "", "Package path (e.g., ./internal/foo)")
	output := flag.String("o", "", "Output file (default: stdout)")
	flag.Parse()

	if *pkgPath == "" {
		fatalf("Usage: gentest -pkg <package-path> [-o output-file]")
	}

	info, err := analyzePackage(*pkgPath)
	if err != nil {
		fatalf("analyze package: %v", err)
	}

	var buf strings.Builder
	if err := testTmpl.Execute(&buf, info); err != nil {
		fatalf("generate template: %v", err)
	}

	result := buf.String()

	if *output != "" {
		if err := os.WriteFile(*output, []byte(result), 0o644); err != nil {
			fatalf("write output: %v", err)
		}
		fmt.Printf("Generated %s\n", *output)
	} else {
		fmt.Println(result)
	}
}

func analyzePackage(pkgPath string) (*PackageInfo, error) {
	absPath, err := filepath.Abs(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, absPath, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse dir: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found in %s", absPath)
	}

	var pkgName string
	var funcs []FuncInfo
	var types []TypeInfo
	importSet := make(map[string]bool)

	for name, pkg := range pkgs {
		pkgName = name
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					fi := parseFunc(d)
					funcs = append(funcs, fi)
				case *ast.GenDecl:
					if d.Tok == token.TYPE {
						for _, spec := range d.Specs {
							ts, ok := spec.(*ast.TypeSpec)
							if !ok {
								continue
							}
							ti := TypeInfo{Name: ts.Name.Name}
							if st, ok := ts.Type.(*ast.StructType); ok {
								for _, field := range st.Fields.List {
									if len(field.Names) > 0 {
										ti.Fields = append(ti.Fields, field.Names[0].Name)
									}
								}
							}
							types = append(types, ti)
						}
					}
					for _, spec := range d.Specs {
						if im, ok := spec.(*ast.ImportSpec); ok {
							imp := strings.Trim(im.Path.Value, `"`)
							if !strings.HasPrefix(imp, "github.com/NAEOS-foundation/naeos/internal") &&
								!strings.HasPrefix(imp, "github.com/NAEOS-foundation/naeos/pkg") {
								continue
							}
							importSet[imp] = true
						}
					}
				}
			}
		}
	}

	funcs = deduplicateFuncs(funcs)
	sort.Slice(funcs, func(i, j int) bool { return funcs[i].Name < funcs[j].Name })
	sort.Slice(types, func(i, j int) bool { return types[i].Name < types[j].Name })

	var imports []string
	for imp := range importSet {
		imports = append(imports, `"`+imp+`"`)
	}
	sort.Strings(imports)

	return &PackageInfo{
		Name:    pkgName,
		Path:    absPath,
		Imports: imports,
		Funcs:   funcs,
		Types:   types,
	}, nil
}

func parseFunc(fn *ast.FuncDecl) FuncInfo {
	fi := FuncInfo{Name: fn.Name.Name}

	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		fi.IsMethod = true
		switch t := fn.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				fi.RecvType = ident.Name
			}
		case *ast.Ident:
			fi.RecvType = t.Name
		}
	}

	if fn.Type.Params != nil {
		for _, p := range fn.Type.Params.List {
			typeStr := exprString(p.Type)
			for range p.Names {
				fi.Params = append(fi.Params, typeStr)
			}
		}
	}

	if fn.Type.Results != nil {
		for _, r := range fn.Type.Results.List {
			typeStr := exprString(r.Type)
			fi.Results = append(fi.Results, typeStr)
		}
	}

	return fi
}

func exprString(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprString(t.X)
	case *ast.SelectorExpr:
		return exprString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + exprString(t.Elt)
		}
		return "[" + exprString(t.Len) + "]" + exprString(t.Elt)
	case *ast.MapType:
		return "map[" + exprString(t.Key) + "]" + exprString(t.Value)
	case *ast.InterfaceType:
		return "any"
	case *ast.FuncType:
		return "func(...)"
	case *ast.BasicLit:
		return t.Value
	default:
		return fmt.Sprintf("%T", e)
	}
}

func deduplicateFuncs(funcs []FuncInfo) []FuncInfo {
	seen := make(map[string]int)
	var result []FuncInfo
	for _, f := range funcs {
		key := f.Name
		if f.IsMethod {
			key = f.RecvType + "_" + f.Name
		}
		if idx, ok := seen[key]; ok {
			result[idx] = f
			continue
		}
		seen[key] = len(result)
		result = append(result, f)
	}

	for i, f := range result {
		if f.IsMethod {
			result[i].Name = f.RecvType + "_" + f.Name
		}
	}
	return result
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
