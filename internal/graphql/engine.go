package graphql

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Schema struct {
	Types   map[string]*TypeDef
	Queries *OperationDef
	Mutations *OperationDef
}

type TypeDef struct {
	Name   string
	Fields map[string]*FieldDef
}

type FieldDef struct {
	Name       string
	Type       string
	Required   bool
	Args       map[string]*ArgDef
	Resolve    Resolver
	IsList     bool
	IsNullable bool
}

type ArgDef struct {
	Name     string
	Type     string
	Required bool
	Default  any
}

type OperationDef struct {
	Fields map[string]*FieldDef
}

type Resolver func(ctx *Context, args map[string]any) (any, error)

type Context struct {
	Request *http.Request
	Schema  *Schema
	Root    any
}

type Request struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]any `json:"variables"`
}

type Response struct {
	Data   any    `json:"data,omitempty"`
	Errors []*GraphQLError `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []Location            `json:"locations,omitempty"`
	Path       []any         `json:"path,omitempty"`
	Extensions map[string]any `json:"extensions,omitempty"`
}

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Executor struct {
	schema *Schema
}

func NewExecutor(schema *Schema) *Executor {
	return &Executor{schema: schema}
}

func (e *Executor) Execute(ctx *Context, query string) *Response {
	ast, errs := ParseQuery(query)
	if len(errs) > 0 {
		return &Response{Errors: errs}
	}

	data := make(map[string]any)
	for _, selection := range ast.Selections {
		result, err := e.resolveSelection(ctx, selection)
		if err != nil {
			return &Response{Errors: []*GraphQLError{{Message: err.Error()}}}
		}
		data[selection.Name] = result
	}

	return &Response{Data: data}
}

func (e *Executor) resolveSelection(ctx *Context, sel *Selection) (any, error) {
	// Try queries first
	if e.schema.Queries != nil {
		if field, ok := e.schema.Queries.Fields[sel.Name]; ok {
			args := e.buildArgs(sel.Arguments, field.Args)
			return field.Resolve(ctx, args)
		}
	}

	// Try mutations
	if e.schema.Mutations != nil {
		if field, ok := e.schema.Mutations.Fields[sel.Name]; ok {
			args := e.buildArgs(sel.Arguments, field.Args)
			return field.Resolve(ctx, args)
		}
	}

	// Try root fields
	if rootMap, ok := ctx.Root.(map[string]any); ok {
		if val, ok := rootMap[sel.Name]; ok {
			return val, nil
		}
	}

	return nil, fmt.Errorf("field '%s' not found", sel.Name)
}

func (e *Executor) buildArgs(arguments map[string]string, argDefs map[string]*ArgDef) map[string]any {
	result := make(map[string]any)
	for name, def := range argDefs {
		if val, ok := arguments[name]; ok {
			result[name] = parseValue(val)
		} else if def.Default != nil {
			result[name] = def.Default
		}
	}
	return result
}

func parseValue(s string) any {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return strings.Trim(s, "\"")
	}
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	if s == "null" {
		return nil
	}
	return s
}

func (e *Executor) Introspect() *Response {
	types := make(map[string]any)
	for name, t := range e.schema.Types {
		fields := make([]map[string]any, 0)
		for _, f := range t.Fields {
			fields = append(fields, map[string]any{
				"name":     f.Name,
				"type":     f.Type,
				"required": f.Required,
			})
		}
		types[name] = map[string]any{
			"name":   name,
			"fields": fields,
		}
	}

	return &Response{
		Data: map[string]any{
			"__schema": map[string]any{
				"types": types,
			},
		},
	}
}

// Simple Query Parser

type QueryAST struct {
	Selections []*Selection
}

type Selection struct {
	Name      string
	Arguments map[string]string
	Children  []*Selection
}

func ParseQuery(query string) (*QueryAST, []*GraphQLError) {
	var errs []*GraphQLError
	ast := &QueryAST{}

	// Remove braces
	query = strings.TrimSpace(query)
	query = strings.TrimPrefix(query, "{")
	query = strings.TrimSuffix(query, "}")
	query = strings.TrimSpace(query)

	if query == "" {
		return ast, nil
	}

	// Tokenize respecting parentheses
	tokens := tokenize(query)

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		sel, err := parseSelection(token)
		if err != nil {
			errs = append(errs, &GraphQLError{Message: err.Error()})
			continue
		}
		ast.Selections = append(ast.Selections, sel)
	}

	return ast, errs
}

func tokenize(query string) []string {
	var tokens []string
	var current strings.Builder
	depth := 0
	inQuote := false

	for i := 0; i < len(query); i++ {
		ch := query[i]

		if ch == '"' {
			inQuote = !inQuote
			current.WriteByte(ch)
			continue
		}

		if inQuote {
			current.WriteByte(ch)
			continue
		}

		if ch == '(' {
			depth++
			current.WriteByte(ch)
			continue
		}

		if ch == ')' {
			depth--
			current.WriteByte(ch)
			continue
		}

		if (ch == ' ' || ch == '\n') && depth == 0 {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteByte(ch)
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

func parseSelection(line string) (*Selection, error) {
	sel := &Selection{
		Arguments: make(map[string]string),
	}

	// Remove trailing comma
	line = strings.TrimRight(line, ",")

	// Check for arguments
	if idx := strings.Index(line, "("); idx != -1 {
		sel.Name = strings.TrimSpace(line[:idx])
		argsStr := line[idx+1:]
		argsStr = strings.TrimSuffix(argsStr, ")")

		args := strings.Split(argsStr, ",")
		for _, arg := range args {
			parts := strings.SplitN(arg, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				sel.Arguments[key] = val
			}
		}
	} else {
		sel.Name = strings.TrimSpace(line)
	}

	if sel.Name == "" {
		return nil, fmt.Errorf("empty selection")
	}

	return sel, nil
}

// HTTP Handler

func Handler(schema *Schema) http.HandlerFunc {
	executor := NewExecutor(schema)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == "GET" {
			introspect := r.URL.Query().Get("introspect")
			if introspect == "true" {
				resp := executor.Introspect()
				json.NewEncoder(w).Encode(resp)
				return
			}
		}

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(&Response{
				Errors: []*GraphQLError{{Message: "method not allowed"}},
			})
			return
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&Response{
				Errors: []*GraphQLError{{Message: "invalid request body"}},
			})
			return
		}

		ctx := &Context{
			Request: r,
			Schema:  schema,
		}

		resp := executor.Execute(ctx, req.Query)
		json.NewEncoder(w).Encode(resp)
	}
}
