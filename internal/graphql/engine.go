package graphql

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	naeoserr "github.com/NAEOS-foundation/naeos/internal/errors"
)

type Schema struct {
	Types     map[string]*TypeDef
	Queries   *OperationDef
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
	Request   *http.Request
	Schema    *Schema
	Root      any
	Variables map[string]any
	path      []any
	errors    []*GraphQLError
	mu        sync.Mutex
}

func (c *Context) AddError(err *GraphQLError) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errors = append(c.errors, err)
}

func (c *Context) Errors() []*GraphQLError {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]*GraphQLError, len(c.errors))
	copy(out, c.errors)
	return out
}

type Request struct {
	Query         string         `json:"query"`
	OperationName string         `json:"operationName"`
	Variables     map[string]any `json:"variables"`
}

type Response struct {
	Data   any             `json:"data,omitempty"`
	Errors []*GraphQLError `json:"errors,omitempty"`
}

type GraphQLError struct {
	Message    string         `json:"message"`
	Locations  []Location     `json:"locations,omitempty"`
	Path       []any          `json:"path,omitempty"`
	Extensions map[string]any `json:"extensions,omitempty"`
}

func (e *GraphQLError) Error() string {
	return e.Message
}

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Executor struct {
	schema     *Schema
	maxDepth   int
	middleware []MiddlewareFunc
	mu         sync.RWMutex
}

type MiddlewareFunc func(ctx *Context, next func(ctx *Context) (*Response, error)) (*Response, error)

func NewExecutor(schema *Schema) *Executor {
	return &Executor{
		schema:   schema,
		maxDepth: 15,
	}
}

func (e *Executor) SetMaxDepth(depth int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.maxDepth = depth
}

func (e *Executor) Use(mw MiddlewareFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.middleware = append(e.middleware, mw)
}

func (e *Executor) Execute(ctx *Context, query string) *Response {
	ast, errs := ParseQuery(query)
	if len(errs) > 0 {
		return &Response{Errors: errs}
	}

	ctx.Schema = e.schema
	ctx.path = []any{}
	ctx.errors = nil

	if ctx.Variables == nil {
		ctx.Variables = make(map[string]any)
	}

	handler := func(c *Context) (*Response, error) {
		data := make(map[string]any)
		for _, selection := range ast.Selections {
			result, err := e.resolveSelection(c, selection, 0)
			if err != nil {
				c.AddError(&GraphQLError{
					Message: err.Error(),
					Path:    c.path,
				})
				continue
			}
			data[selection.Name] = result
		}
		return &Response{Data: data}, nil
	}

	for i := len(e.middleware) - 1; i >= 0; i-- {
		mw := e.middleware[i]
		nextHandler := handler
		handler = func(c *Context) (*Response, error) {
			return mw(c, nextHandler)
		}
	}

	resp, err := handler(ctx)
	if err != nil {
		ctx.AddError(&GraphQLError{Message: err.Error()})
	}

	resp.Errors = ctx.Errors()
	return resp
}

func (e *Executor) resolveSelection(ctx *Context, sel *Selection, depth int) (any, error) {
	if depth > e.maxDepth {
		return nil, naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("query depth exceeds maximum of %d", e.maxDepth))
	}

	if e.schema.Queries != nil {
		if field, ok := e.schema.Queries.Fields[sel.Name]; ok {
			args := e.buildArgs(sel.Arguments, field.Args, ctx.Variables)
			result, err := field.Resolve(ctx, args)
			if err != nil {
				return nil, err
			}
			if len(sel.Children) > 0 {
				return e.resolveChildren(ctx, result, sel.Children, depth+1)
			}
			return result, nil
		}
	}

	if e.schema.Mutations != nil {
		if field, ok := e.schema.Mutations.Fields[sel.Name]; ok {
			args := e.buildArgs(sel.Arguments, field.Args, ctx.Variables)
			result, err := field.Resolve(ctx, args)
			if err != nil {
				return nil, err
			}
			if len(sel.Children) > 0 {
				return e.resolveChildren(ctx, result, sel.Children, depth+1)
			}
			return result, nil
		}
	}

	if rootMap, ok := ctx.Root.(map[string]any); ok {
		if val, ok := rootMap[sel.Name]; ok {
			if len(sel.Children) > 0 {
				return e.resolveChildren(ctx, val, sel.Children, depth+1)
			}
			return val, nil
		}
	}

	return nil, naeoserr.New(naeoserr.ErrNotFound, fmt.Sprintf("field '%s' not found", sel.Name))
}

func (e *Executor) resolveChildren(ctx *Context, parent any, children []*Selection, depth int) (any, error) {
	if depth > e.maxDepth {
		return nil, naeoserr.New(naeoserr.ErrValidation, fmt.Sprintf("query depth exceeds maximum of %d", e.maxDepth))
	}

	parentMap, ok := toMap(parent)
	if !ok {
		return nil, naeoserr.New(naeoserr.ErrValidation, "cannot resolve sub-fields on non-object type")
	}

	result := make(map[string]any)
	for _, child := range children {
		val, ok := parentMap[child.Name]
		if !ok {
			val = nil
		}
		if len(child.Children) > 0 && val != nil {
			var err error
			val, err = e.resolveChildren(ctx, val, child.Children, depth+1)
			if err != nil {
				ctx.AddError(&GraphQLError{
					Message: fmt.Sprintf("field '%s': %s", child.Name, err.Error()),
				})
				val = nil
			}
		}
		result[child.Name] = val
	}
	return result, nil
}

func (e *Executor) buildArgs(arguments map[string]string, argDefs map[string]*ArgDef, variables map[string]any) map[string]any {
	result := make(map[string]any)
	for name, def := range argDefs {
		if val, ok := arguments[name]; ok {
			if strings.HasPrefix(val, "$") {
				varName := strings.TrimPrefix(val, "$")
				if v, exists := variables[varName]; exists {
					result[name] = v
					continue
				}
			}
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
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	return s
}

func toMap(v any) (map[string]any, bool) {
	switch m := v.(type) {
	case map[string]any:
		return m, true
	case map[string]string:
		result := make(map[string]any, len(m))
		for k, val := range m {
			result[k] = val
		}
		return result, true
	}
	return nil, false
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

	queryFields := make([]map[string]any, 0)
	if e.schema.Queries != nil {
		for _, f := range e.schema.Queries.Fields {
			queryFields = append(queryFields, map[string]any{
				"name": f.Name,
				"type": f.Type,
			})
		}
	}

	return &Response{
		Data: map[string]any{
			"__schema": map[string]any{
				"types":     types,
				"queryType": queryFields,
			},
		},
	}
}

// Query Parser

type QueryAST struct {
	Selections []*Selection
	Fragments  map[string]*FragmentDef
}

type Selection struct {
	Name      string
	Arguments map[string]string
	Children  []*Selection
}

type FragmentDef struct {
	Name       string
	OnType     string
	Selections []*Selection
}

func ParseQuery(query string) (*QueryAST, []*GraphQLError) {
	var errs []*GraphQLError
	ast := &QueryAST{
		Fragments: make(map[string]*FragmentDef),
	}

	query = strings.TrimSpace(query)
	query = strings.TrimPrefix(query, "{")
	query = strings.TrimSuffix(query, "}")
	query = strings.TrimSpace(query)

	if query == "" {
		return ast, nil
	}

	tokens := tokenize(query)

	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		if strings.HasPrefix(token, "fragment ") {
			frag, err := parseFragment(token)
			if err != nil {
				errs = append(errs, &GraphQLError{Message: err.Error()})
				continue
			}
			ast.Fragments[frag.Name] = frag
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

func parseFragment(token string) (*FragmentDef, error) {
	token = strings.TrimPrefix(token, "fragment ")
	parts := strings.SplitN(token, "{", 2)
	if len(parts) != 2 {
		return nil, naeoserr.New(naeoserr.ErrParse, "invalid fragment syntax")
	}
	header := strings.TrimSpace(parts[0])
	onParts := strings.Split(header, " on ")
	if len(onParts) != 2 {
		return nil, naeoserr.New(naeoserr.ErrParse, "fragment must specify type (fragment X on Type)")
	}
	name := strings.TrimSpace(onParts[0])
	onType := strings.TrimSpace(onParts[1])

	inner := strings.TrimSuffix(parts[1], "}")
	inner = strings.TrimSpace(inner)

	sels, err := parseNestedSelections(inner)
	if err != nil {
		return nil, err
	}

	return &FragmentDef{
		Name:       name,
		OnType:     onType,
		Selections: sels,
	}, nil
}

func parseNestedSelections(inner string) ([]*Selection, error) {
	var selections []*Selection
	tokens := tokenize(inner)
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		sel, err := parseSelection(token)
		if err != nil {
			return nil, err
		}
		selections = append(selections, sel)
	}
	return selections, nil
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

		if ch == '{' {
			depth++
			if depth == 1 {
				if current.Len() > 0 {
					current.WriteByte(ch)
					continue
				}
				current.WriteByte(ch)
				continue
			}
			current.WriteByte(ch)
			continue
		}

		if ch == '}' {
			depth--
			current.WriteByte(ch)
			if depth == 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}

		if (ch == ' ' || ch == '\n' || ch == '\t') && depth == 0 {
			remaining := query[i+1:]
			nextNonSpace := byte(0)
			for j := 0; j < len(remaining); j++ {
				if remaining[j] != ' ' && remaining[j] != '\n' && remaining[j] != '\t' {
					nextNonSpace = remaining[j]
					break
				}
			}
			if current.Len() > 0 && nextNonSpace == '{' {
				current.WriteByte(ch)
				continue
			}
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

	line = strings.TrimRight(line, ",")
	line = strings.TrimSpace(line)

	if line == "" {
		return nil, naeoserr.New(naeoserr.ErrParse, "empty selection")
	}

	bracketIdx := strings.Index(line, "{")
	parenIdx := strings.Index(line, "(")

	if parenIdx != -1 && (bracketIdx == -1 || parenIdx < bracketIdx) {
		sel.Name = strings.TrimSpace(line[:parenIdx])

		depth := 0
		endParen := -1
		for i := parenIdx; i < len(line); i++ {
			if line[i] == '(' {
				depth++
			} else if line[i] == ')' {
				depth--
				if depth == 0 {
					endParen = i
					break
				}
			}
		}
		if endParen == -1 {
			endParen = strings.Index(line[parenIdx:], ")")
			if endParen == -1 {
				endParen = len(line)
			} else {
				endParen += parenIdx
			}
		}

		argsStr := line[parenIdx+1 : endParen]
		args := splitArgs(argsStr)
		for _, arg := range args {
			parts := strings.SplitN(arg, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				sel.Arguments[key] = val
			}
		}

		rest := strings.TrimSpace(line[endParen+1:])
		if strings.HasPrefix(rest, "{") {
			inner := rest[1:]
			inner = strings.TrimSuffix(inner, "}")
			parsedChildren, err := parseNestedSelections(inner)
			if err != nil {
				return nil, err
			}
			sel.Children = parsedChildren
		}
	} else if bracketIdx != -1 {
		header := strings.TrimSpace(line[:bracketIdx])
		inner := line[bracketIdx+1:]
		inner = strings.TrimSuffix(inner, "}")

		sel.Name = extractName(header)

		parsedChildren, err := parseNestedSelections(inner)
		if err != nil {
			return nil, err
		}
		sel.Children = parsedChildren
	} else {
		sel.Name = strings.TrimSpace(line)
	}

	if sel.Name == "" {
		return nil, naeoserr.New(naeoserr.ErrParse, "empty selection")
	}

	return sel, nil
}

func extractName(header string) string {
	parenIdx := strings.Index(header, "(")
	if parenIdx != -1 {
		return strings.TrimSpace(header[:parenIdx])
	}
	return strings.TrimSpace(header)
}

func splitArgs(argsStr string) []string {
	var args []string
	var current strings.Builder
	depth := 0
	inQuote := false

	for i := 0; i < len(argsStr); i++ {
		ch := argsStr[i]
		if ch == '"' {
			inQuote = !inQuote
		}
		if inQuote {
			current.WriteByte(ch)
			continue
		}
		if ch == '(' {
			depth++
		}
		if ch == ')' {
			depth--
		}
		if ch == ',' && depth == 0 {
			args = append(args, current.String())
			current.Reset()
			continue
		}
		current.WriteByte(ch)
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
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
				_ = json.NewEncoder(w).Encode(resp)
				return
			}
		}

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(&Response{
				Errors: []*GraphQLError{{Message: "method not allowed"}},
			})
			return
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(&Response{
				Errors: []*GraphQLError{{Message: "invalid request body"}},
			})
			return
		}

		ctx := &Context{
			Request:   r,
			Schema:    schema,
			Variables: req.Variables,
		}

		resp := executor.Execute(ctx, req.Query)
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// Helper for building resolvers that return map data
func MapResolver(data map[string]any) Resolver {
	return func(ctx *Context, args map[string]any) (any, error) {
		return data, nil
	}
}

// Helper for building list resolvers
func ListResolver(items []any) Resolver {
	return func(ctx *Context, args map[string]any) (any, error) {
		return items, nil
	}
}

