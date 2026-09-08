package signing

import (
	"bytes"
	"encoding/json"
)

func encodeJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
