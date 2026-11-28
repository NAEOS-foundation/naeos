// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package identitycontract

import "testing"

func TestPrincipalValidate(t *testing.T) {
	cases := []struct {
		name string
		p    Principal
		want bool
	}{
		{"valid", Principal{Issuer: "https://issuer.example", Subject: "subject-123"}, true},
		{"missing issuer", Principal{Subject: "subject-123"}, false},
		{"missing subject", Principal{Issuer: "https://issuer.example"}, false},
		{"blank issuer", Principal{Issuer: " ", Subject: "subject-123"}, false},
		{"blank subject", Principal{Issuer: "https://issuer.example", Subject: "  "}, false},
		{"issuer surrounding whitespace", Principal{Issuer: " https://issuer.example", Subject: "subject-123"}, false},
		{"subject surrounding whitespace", Principal{Issuer: "https://issuer.example", Subject: " subject-123"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.p.Validate()
			if (err == nil) != tc.want {
				t.Fatalf("Validate() error = %v; valid expected=%v", err, tc.want)
			}
		})
	}
}

func TestPrincipalKeepsIssuerQualifiedIdentityExact(t *testing.T) {
	p := Principal{Issuer: "https://issuer-a.example", Subject: "user-42"}
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
	if p.Issuer != "https://issuer-a.example" || p.Subject != "user-42" {
		t.Fatalf("principal mutated: %+v", p)
	}
}
