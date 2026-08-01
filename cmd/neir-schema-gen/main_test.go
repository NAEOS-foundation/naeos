package main

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/neir/model/deployment"
)

func TestParseJSONTag(t *testing.T) {
	t.Parallel()

	type tagged struct {
		A string `json:"simple"`
		B string `json:"opt,omitempty"`
		C string `json:"-"`
		D string `json:"field,omitempty,inline"`
		E string
	}

	tests := []struct {
		name     string
		field    string
		wantName string
		wantOpts map[string]bool
	}{
		{name: "simple", field: "A", wantName: "simple", wantOpts: map[string]bool{}},
		{name: "omitempty", field: "B", wantName: "opt", wantOpts: map[string]bool{"omitempty": true}},
		{name: "dash", field: "C", wantName: "-", wantOpts: map[string]bool{}},
		{name: "multi", field: "D", wantName: "field", wantOpts: map[string]bool{"omitempty": true, "inline": true}},
		{name: "empty", field: "E", wantName: "", wantOpts: nil},
	}

	typ := reflect.TypeOf(tagged{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _ := typ.FieldByName(tt.field)
			gotName, gotOpts := parseJSONTag(f.Tag.Get("json"))
			if gotName != tt.wantName {
				t.Errorf("name = %q, want %q", gotName, tt.wantName)
			}
			if len(gotOpts) != len(tt.wantOpts) {
				t.Errorf("opts = %v, want %v", gotOpts, tt.wantOpts)
			}
			for k, v := range tt.wantOpts {
				if gotOpts[k] != v {
					t.Errorf("opts[%q] = %v, want %v", k, gotOpts[k], v)
				}
			}
		})
	}
}

func TestJSONType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		typ  reflect.Type
		want string
	}{
		{reflect.TypeOf(""), "string"},
		{reflect.TypeOf(false), "boolean"},
		{reflect.TypeOf(0), "number"},
		{reflect.TypeOf(int64(0)), "number"},
		{reflect.TypeOf(float64(0)), "number"},
		{reflect.TypeOf([]byte(nil)), "string"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := jsonType(tt.typ)
			if got != tt.want {
				t.Errorf("jsonType(%v) = %q, want %q", tt.typ, got, tt.want)
			}
		})
	}
}

func TestTypeName(t *testing.T) {
	t.Parallel()

	got := typeName(reflect.TypeOf(""))
	if !strings.Contains(got, "string") {
		t.Errorf("typeName(string) = %q, want to contain 'string'", got)
	}
}

func TestKnownEnums(t *testing.T) {
	t.Parallel()

	enums := knownEnums()
	expected := []string{"architecture.Pattern", "service.ServiceKind", "api.Protocol"}
	for _, key := range expected {
		if _, ok := enums[key]; !ok {
			t.Errorf("knownEnums() missing key %q", key)
		}
	}
	if len(enums) < 8 {
		t.Errorf("knownEnums() returned %d entries, want >= 8", len(enums))
	}
}

func TestGeneratorRequiredRoot(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	required := g.requiredRoot()
	if len(required) != 2 {
		t.Fatalf("len(requiredRoot()) = %d, want 2", len(required))
	}
	if required[0] != "project" {
		t.Errorf("requiredRoot()[0] = %q, want %q", required[0], "project")
	}
	if required[1] != "modules" {
		t.Errorf("requiredRoot()[1] = %q, want %q", required[1], "modules")
	}
}

func TestGeneratorPrimitiveSchema(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	schema := g.primitiveSchema(reflect.TypeOf(""), false)
	if schema["type"] != "string" {
		t.Errorf("primitiveSchema(string) type = %v, want 'string'", schema["type"])
	}
}

func TestFieldSchema(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	t.Run("string", func(t *testing.T) {
		s := g.fieldSchema(reflect.TypeOf(""), false)
		if s["type"] != "string" {
			t.Errorf("string schema type = %v", s["type"])
		}
	})

	t.Run("pointer", func(t *testing.T) {
		s := g.fieldSchema(reflect.TypeOf((*string)(nil)), false)
		if s["type"] != "string" {
			t.Errorf("pointer schema type = %v", s["type"])
		}
	})
}

func TestArraySchema(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	schema := g.arraySchema(reflect.TypeOf(""))
	if schema["type"] != "array" {
		t.Errorf("arraySchema type = %v, want 'array'", schema["type"])
	}
	items, ok := schema["items"].(map[string]any)
	if !ok {
		t.Fatal("arraySchema items is not a map")
	}
	if items["type"] != "string" {
		t.Errorf("arraySchema items type = %v, want 'string'", items["type"])
	}
}

func TestMapSchema(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	schema := g.mapSchema(reflect.TypeOf(map[string]int{}))
	if schema["type"] != "object" {
		t.Errorf("mapSchema type = %v, want 'object'", schema["type"])
	}
}

func TestEnumForField(t *testing.T) {
	t.Parallel()

	enums := knownEnums()

	type MyStruct struct {
		Strategy string `json:"strategy"`
	}

	ft := reflect.TypeOf(MyStruct{})
	f, _ := ft.FieldByName("Strategy")

	vals, ok := enumForField(ft, f, enums)
	if ok {
		t.Errorf("enumForField should not match for string field, got %v", vals)
	}
}

func TestFieldSchema_Slice(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	schema := g.fieldSchema(reflect.TypeOf([]string{}), false)
	if schema["type"] != "array" {
		t.Errorf("slice schema type = %v, want 'array'", schema["type"])
	}
}

func TestGenerate_Unexported(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	type Inner struct {
		Exported   string `json:"exported"`
		unexported string
		Hidden     string `json:"-"`
		NoTag      string
	}

	props := g.generate(reflect.TypeOf(Inner{}), false)
	if _, ok := props["exported"]; !ok {
		t.Error("expected 'exported' property")
	}
	if _, ok := props["unexported"]; ok {
		t.Error("unexpected 'unexported' property")
	}
	if _, ok := props["hidden"]; ok {
		t.Error("unexpected 'hidden' property")
	}
	if _, ok := props["NoTag"]; ok {
		t.Error("unexpected 'NoTag' property")
	}
}

func TestStructSchema_ExtraDefinitions(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	type Inner struct {
		X string `json:"x"`
	}
	g.definitions["Inner"] = map[string]any{"extra": "field"}

	props := g.generate(reflect.TypeOf(Inner{}), false)
	if _, ok := props["extra"]; !ok {
		t.Error("expected 'extra' property from definitions")
	}
}

func TestPrimitiveSchema_WithEnum(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	schema := g.primitiveSchema(reflect.TypeOf(""), false)
	if schema["type"] != "string" {
		t.Errorf("string schema type = %v, want 'string'", schema["type"])
	}
}

func TestStructSchema_EnumField(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	type Inner struct {
		Kind string `json:"kind"`
	}

	schema := g.structSchema(reflect.TypeOf(Inner{}), false)
	props, ok := schema["$ref"].(string)
	if !ok {
		t.Fatal("expected $ref")
	}
	if !strings.Contains(props, "Inner") {
		t.Errorf("$ref = %q, want 'Inner'", props)
	}
}

func TestGenerate_EmptyTagField(t *testing.T) {
	t.Parallel()

	type Inner struct {
		X string `json:""`
	}

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	props := g.generate(reflect.TypeOf(Inner{}), false)
	if _, ok := props["X"]; ok {
		t.Error("expected no property for empty json tag field")
	}
}

func TestJSONType_TimeFromPkg(t *testing.T) {
	t.Parallel()

	got := jsonType(reflect.TypeOf(time.Time{}))
	if got != "string" {
		t.Errorf("jsonType(time.Time) = %q, want 'string'", got)
	}
}

func TestStructSchema_Time(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	schema := g.structSchema(reflect.TypeOf(time.Time{}), false)
	if schema["type"] != "string" {
		t.Errorf("time.Time schema type = %v, want 'string'", schema["type"])
	}
	if schema["format"] != "date-time" {
		t.Errorf("time.Time format = %v, want 'date-time'", schema["format"])
	}
}

func TestStructSchema_Reused(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	type Inner struct {
		X string `json:"x"`
	}
	innerType := reflect.TypeOf(Inner{})

	s1 := g.structSchema(innerType, false)
	if _, ok := s1["$ref"]; !ok {
		t.Error("expected $ref for struct schema")
	}

	s2 := g.structSchema(innerType, false)
	ref, ok := s2["$ref"]
	if !ok {
		t.Error("expected $ref for reused struct")
	}
	if !strings.HasSuffix(ref.(string), ".Inner") {
		t.Errorf("$ref = %q, want suffix .Inner", ref)
	}
}

func TestStructSchema_Optional(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	type Inner struct {
		X string `json:"x"`
	}
	innerType := reflect.TypeOf(Inner{})

	schema := g.structSchema(innerType, true)
	if _, ok := schema["$ref"]; !ok {
		t.Error("expected $ref for optional struct schema")
	}
}

func TestGenerate(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	type MyStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age,omitempty"`
	}

	props := g.generate(reflect.TypeOf(MyStruct{}), false)
	if props["name"] == nil {
		t.Error("expected 'name' property")
	}
	if props["age"] == nil {
		t.Error("expected 'age' property")
	}
}

func TestFieldSchema_Map(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	schema := g.fieldSchema(reflect.TypeOf(map[string]string{}), false)
	if schema["type"] != "object" {
		t.Errorf("map schema type = %v, want 'object'", schema["type"])
	}
}

func TestFieldSchema_Struct(t *testing.T) {
	t.Parallel()

	g := &generator{
		definitions: make(map[string]map[string]any),
		visited:     make(map[string]bool),
		enums:       knownEnums(),
	}

	type Inner struct {
		V string `json:"v"`
	}
	schema := g.fieldSchema(reflect.TypeOf(Inner{}), false)
	if _, ok := schema["$ref"]; !ok {
		t.Error("expected $ref for struct field schema")
	}
}

func TestJSONType_Uint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		typ  reflect.Type
		want string
	}{
		{reflect.TypeOf(uint(0)), "number"},
		{reflect.TypeOf(uint8(0)), "number"},
		{reflect.TypeOf(uint16(0)), "number"},
		{reflect.TypeOf(uint32(0)), "number"},
		{reflect.TypeOf(uint64(0)), "number"},
		{reflect.TypeOf(float32(0)), "number"},
	}
	for _, tt := range tests {
		t.Run(tt.typ.String(), func(t *testing.T) {
			got := jsonType(tt.typ)
			if got != tt.want {
				t.Errorf("jsonType(%v) = %q, want %q", tt.typ, got, tt.want)
			}
		})
	}
}

func TestEnumForField_Pointer(t *testing.T) {
	t.Parallel()

	enums := knownEnums()
	type MyStruct struct {
		P *deployment.Strategy `json:"strategy"`
	}

	ft := reflect.TypeOf(MyStruct{})
	f, _ := ft.FieldByName("P")
	vals, ok := enumForField(ft, f, enums)
	if !ok {
		t.Error("expected enum for *deployment.Strategy")
	}
	if len(vals) == 0 {
		t.Error("expected non-empty enum values")
	}
}

func TestEnumForField_Slice(t *testing.T) {
	t.Parallel()

	enums := knownEnums()
	type MyStruct struct {
		Items []string `json:"items"`
	}

	ft := reflect.TypeOf(MyStruct{})
	f, _ := ft.FieldByName("Items")
	_, ok := enumForField(ft, f, enums)
	if ok {
		t.Error("expected no enum for slice field")
	}
}
