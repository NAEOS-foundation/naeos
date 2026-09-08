package sbom

import (
	"bytes"
	"encoding/json"
)

// EncodeJSON marshals v to indented JSON bytes.
func EncodeJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeJSON unmarshals data into v.
func DecodeJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
