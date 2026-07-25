package auth

import (
	"testing"
)

func FuzzParseLDAPResult(f *testing.F) {
	// Valid LDAP SEARCH response (minimal)
	f.Add([]byte{
		0x30, 0x0e, // SEQUENCE of length 14
		0x64, 0x0c, // SearchResultEntry of length 12
		0x30, 0x0a, // SEQUENCE of length 10
		0x04, 0x04, 0x75, 0x69, 0x64, 0x3d, // OCTET STRING "uid="
		0x31, 0x02, // SET of length 2
		0x04, 0x00, // empty attribute value
	})
	f.Add([]byte{0x30, 0x00})                                              // empty SEQUENCE
	f.Add([]byte{0x30, 0x02, 0x04, 0x00})                                 // SEQUENCE with empty string
	f.Add([]byte{0x30, 0x05, 0x04, 0x03, 0x66, 0x6f, 0x6f})              // SEQUENCE with "foo"
	f.Add([]byte{0x30, 0x07, 0x64, 0x05, 0x30, 0x03, 0x04, 0x01, 0x61}) // SEQUENCE > SearchResultEntry > string "a"
	f.Add([]byte{})                                                        // empty
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})                                 // invalid tags

	f.Fuzz(func(t *testing.T, data []byte) {
		result := parseLDAPResult(data)
		// Should never panic; result may be empty or partial
		_ = result
	})
}

func FuzzParseLDAPSequence(f *testing.F) {
	f.Add([]byte{0x30, 0x00})                              // empty SEQUENCE
	f.Add([]byte{0x30, 0x05, 0x04, 0x03, 0x61, 0x62, 0x63}) // SEQUENCE with "abc"
	f.Add([]byte{0x64, 0x03, 0x04, 0x01, 0x78})           // SearchResultEntry with "x"
	f.Add([]byte{0x31, 0x02, 0x04, 0x00})                 // SET with empty string
	f.Add([]byte{})                                        // empty
	f.Add([]byte{0x30})                                    // truncated tag
	f.Add([]byte{0x30, 0x01})                              // truncated length
	f.Add([]byte{0x30, 0x10, 0x04, 0x0f})                 // body shorter than declared length

	f.Fuzz(func(t *testing.T, data []byte) {
		_, items := parseLDAPSequence(data)
		_ = items
	})
}

func FuzzParseLDAPString(f *testing.F) {
	f.Add([]byte{0x04, 0x03, 0x66, 0x6f, 0x6f}) // OCTET STRING "foo"
	f.Add([]byte{0x04, 0x00})                    // empty string
	f.Add([]byte{})                              // empty
	f.Add([]byte{0x04})                          // truncated
	f.Add([]byte{0x04, 0x05, 0x61})             // body shorter than declared length
	f.Add([]byte{0x04, 0xff})                   // max single-byte length (no body)
	f.Add(make([]byte, 256))                     // long slice with byte(0) length

	f.Fuzz(func(t *testing.T, data []byte) {
		result := parseLDAPString(data)
		_ = result
	})
}

func FuzzSAMLParserResponse(f *testing.F) {
	f.Add(`<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
  xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">
  <samlp:Status><samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/></samlp:Status>
  <saml:Assertion ID="_abc123" IssueInstant="2024-01-01T00:00:00Z">
    <saml:Issuer>https://idp.example.com</saml:Issuer>
    <saml:Subject><saml:NameID>user@example.com</saml:NameID></saml:Subject>
  </saml:Assertion>
</samlp:Response>`)
	f.Add(`invalid xml`)
	f.Add(``)
	f.Add(`<Response><Status><StatusCode Value="Failure"/></Status></Response>`)
	f.Add(`<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol">
  <saml:Assertion>
    <saml:Subject><saml:NameID>test</saml:NameID></saml:Subject>
    <saml:AttributeStatement>
      <saml:Attribute Name="email">
        <saml:AttributeValue>test@example.com</saml:AttributeValue>
      </saml:Attribute>
    </saml:AttributeStatement>
  </saml:Assertion>
</samlp:Response>`)
	// base64-encoded minimal valid SAML response
	f.Add("PD94bWwgdmVyc2lvbj0iMS4wIj8+PFJlc3BvbnNlPjwvUmVzcG9uc2U+")

	p := &SAMLProvider{}

	f.Fuzz(func(t *testing.T, raw string) {
		user, err := p.ParseResponse(raw)
		if err != nil {
			_ = err
			return
		}
		_ = user
	})
}
