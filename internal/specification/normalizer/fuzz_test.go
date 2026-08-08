package normalizer

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/NAEOS-foundation/naeos/internal/specification/parser"
)

func FuzzNormalize(f *testing.F) {
	f.Add(`project: test`)
	f.Add(`project: test
modules:
  - name: core
    path: ./core`)
	f.Add(`project: test
services:
  - name: api
    port: 8080
    endpoints:
      - method: GET
        path: /health`)
	f.Add(`project: full
modules:
  - name: auth
    dependencies: [core]
services:
  - name: gateway
    kind: http
    port: 443
architecture:
  pattern: hexagonal
deployment:
  strategy: rolling
testing:
  strategy: unit`)
	f.Add(`{invalid`)
	f.Add(``)
	f.Add(`project: ""`)
	f.Add(`modules: []`)
	f.Add(`services: []`)

	f.Fuzz(func(t *testing.T, input string) {
		p := parser.NewParser(".")
		doc, err := p.Parse(input)
		if err != nil {
			return
		}
		if doc == nil {
			return
		}

		normalizer := NewNormalizer()
		result, err := normalizer.Normalize(doc)
		if err != nil {
			return
		}

		if result == nil {
			t.Fatalf("result should not be nil on success")
		}
		if result.Values == nil {
			t.Error("result values should not be nil")
		}
	})
}

func FuzzNormalizeRaw(f *testing.F) {
	f.Add(`project: test`)
	f.Add(`project: test
modules:
  - name: core
    path: ./core`)
	f.Add(`{invalid}`)
	f.Add(`key: value`)
	f.Add(`project: test
architecture:
  pattern: hexagonal`)
	f.Add(`modules: "not-an-array"`)
	f.Add(`services: "not-an-array"`)

	f.Fuzz(func(t *testing.T, input string) {
		var data map[string]any
		if err := yaml.Unmarshal([]byte(input), &data); err != nil {
			return
		}
		if data == nil {
			return
		}

		result, err := NormalizeRaw(data)
		if err != nil {
			return
		}

		if result == nil {
			t.Fatalf("result should not be nil on success")
		}
		if result.Values == nil {
			t.Error("result values should not be nil")
		}
	})
}

func FuzzFlatten(f *testing.F) {
	f.Add(`project: test`)
	f.Add(`top:
  mid:
    leaf: value`)
	f.Add(`a:
  b:
    c: deep`)
	f.Add(`key1: val1
key2: val2`)
	f.Add(`{}`)

	f.Fuzz(func(t *testing.T, input string) {
		var data map[string]any
		if err := yaml.Unmarshal([]byte(input), &data); err != nil {
			return
		}
		if data == nil {
			return
		}

		flat := Flatten(data)
		if flat == nil {
			t.Error("flatten result should not be nil")
			return
		}

		unflat := Unflatten(flat)
		if unflat == nil {
			t.Error("unflatten result should not be nil")
		}
	})
}
