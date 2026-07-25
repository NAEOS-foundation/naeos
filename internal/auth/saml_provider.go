package auth

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

type SAMLProvider struct {
	config      SSOConfig
	certificate *x509.Certificate
}

type SAMLResponse struct {
	XMLName    xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol Response"`
	Assertion  *SAMLAssertion `xml:"Assertion"`
	Status     *SAMLStatus    `xml:"Status"`
}

type SAMLAssertion struct {
	XMLName            xml.Name       `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
	ID                 string         `xml:"ID,attr"`
	IssueInstant       string         `xml:"IssueInstant,attr"`
	Issuer             string         `xml:"Issuer"`
	Subject            *SAMLSubject   `xml:"Subject"`
	Conditions         *SAMLConditions `xml:"Conditions"`
	AttributeStatement *SAMLAttributeStatement `xml:"AttributeStatement"`
	AuthnStatement     *SAMLAuthnStatement    `xml:"AuthnStatement"`
}

type SAMLSubject struct {
	NameID   string `xml:"NameID"`
	SubjectConfirmation *SAMLSubjectConfirmation `xml:"SubjectConfirmation"`
}

type SAMLSubjectConfirmation struct {
	Method string `xml:"Method,attr"`
}

type SAMLConditions struct {
	NotBefore    string `xml:"NotBefore,attr"`
	NotOnOrAfter string `xml:"NotOnOrAfter,attr"`
}

type SAMLAttributeStatement struct {
	Attributes []SAMLAttribute `xml:"Attribute"`
}

type SAMLAttribute struct {
	Name           string   `xml:"Name,attr"`
	FriendlyName   string   `xml:"FriendlyName,attr,omitempty"`
	AttributeValues []SAMLAttributeValue `xml:"AttributeValue"`
}

type SAMLAttributeValue struct {
	Value string `xml:",innerxml"`
}

type SAMLStatus struct {
	StatusCode *SAMLStatusCode `xml:"StatusCode"`
}

type SAMLStatusCode struct {
	Value string `xml:"Value,attr"`
}

type SAMLAuthnStatement struct {
	AuthnInstant string `xml:"AuthnInstant,attr"`
	SessionIndex string `xml:"SessionIndex,attr,omitempty"`
}

func NewSAMLProvider(cfg SSOConfig) *SAMLProvider {
	return &SAMLProvider{config: cfg}
}

func (p *SAMLProvider) Type() ProviderType { return ProviderSAML }

func (p *SAMLProvider) Name() string { return p.config.Name }

func (p *SAMLProvider) Config() SSOConfig { return p.config }

func (p *SAMLProvider) Validate() error {
	if p.config.Name == "" {
		return fmt.Errorf("name is required")
	}
	if p.config.SSOURL == "" {
		return fmt.Errorf("sso_url is required")
	}
	if p.config.EntityID == "" {
		return fmt.Errorf("entity_id is required")
	}
	if p.config.CertFile != "" {
		cert, err := loadCertFile(p.config.CertFile)
		if err != nil {
			return fmt.Errorf("load cert: %w", err)
		}
		p.certificate = cert
	}
	return nil
}

func (p *SAMLProvider) ParseResponse(samlResponse string) (*OAuth2User, error) {
	decoded, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		decoded = []byte(samlResponse)
	}

	var resp SAMLResponse
	if err := xml.Unmarshal(decoded, &resp); err != nil {
		return nil, fmt.Errorf("parse SAML response: %w", err)
	}

	if resp.Status != nil && resp.Status.StatusCode != nil {
		if !strings.HasSuffix(resp.Status.StatusCode.Value, "Success") {
			return nil, fmt.Errorf("SAML response status: %s", resp.Status.StatusCode.Value)
		}
	}

	if resp.Assertion == nil {
		return nil, fmt.Errorf("SAML response missing assertion")
	}

	a := resp.Assertion

	user := &OAuth2User{
		ID:   a.Subject.NameID,
		Name: a.Subject.NameID,
	}

	if a.AttributeStatement != nil {
		for _, attr := range a.AttributeStatement.Attributes {
			for _, av := range attr.AttributeValues {
				val := strings.TrimSpace(av.Value)
				switch strings.ToLower(attr.Name) {
				case "email", "mail", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress":
					user.Email = val
				case "displayname", "cn", "name", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name":
					user.Name = val
				case "uid", "username", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier":
					user.ID = val
				}
			}
		}
	}

	return user, nil
}

func (p *SAMLProvider) GetAuthorizationURL(state string) string {
	return p.config.SSOURL
}

func (p *SAMLProvider) ExchangeCode(code string) (*OAuth2Token, error) {
	return nil, fmt.Errorf("SAML does not support code exchange; use ParseResponse instead")
}

func (p *SAMLProvider) GetUser(token *OAuth2Token) (*OAuth2User, error) {
	return nil, fmt.Errorf("SAML does not support GetUser; use ParseResponse instead")
}

func loadCertFile(path string) (*x509.Certificate, error) {
	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("no PEM data in %s", path)
	}
	return x509.ParseCertificate(block.Bytes)
}
