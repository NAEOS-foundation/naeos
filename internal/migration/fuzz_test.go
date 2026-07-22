package migration

import (
	"testing"
)

func FuzzParseVersion(f *testing.F) {
	f.Add("0.1.0")
	f.Add("0.3.0")
	f.Add("1.0.0")
	f.Add("invalid")
	f.Add("")
	f.Add("1.2.3.4")
	f.Add("0.0")
	f.Add("-1.0.0")
	f.Add("abc.def.ghi")
	f.Add("999.999.999")

	f.Fuzz(func(t *testing.T, input string) {
		v, err := ParseVersion(input)
		if err != nil {
			return
		}
		if v.Major < 0 {
			t.Error("major should not be negative")
		}
		if v.Minor < 0 {
			t.Error("minor should not be negative")
		}
		if v.Patch < 0 {
			t.Error("patch should not be negative")
		}
		roundTrip := v.String()
		v2, err := ParseVersion(roundTrip)
		if err != nil {
			t.Fatalf("round-trip parse failed: %v", err)
		}
		if v != v2 {
			t.Errorf("round-trip mismatch: %v != %v", v, v2)
		}
	})
}
