package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type LDAPProvider struct {
	config SSOConfig
}

type ldapConn struct {
	conn net.Conn
}

func NewLDAPProvider(cfg SSOConfig) *LDAPProvider {
	return &LDAPProvider{config: cfg}
}

func (p *LDAPProvider) Type() ProviderType { return ProviderLDAP }

func (p *LDAPProvider) Name() string { return p.config.Name }

func (p *LDAPProvider) Config() SSOConfig { return p.config }

func (p *LDAPProvider) Validate() error {
	if p.config.Name == "" {
		return fmt.Errorf("name is required")
	}
	if p.config.Host == "" {
		return fmt.Errorf("host is required")
	}
	if p.config.Port == 0 {
		p.config.Port = 389
	}
	if p.config.BaseDN == "" {
		return fmt.Errorf("base_dn is required")
	}
	if p.config.UserFilter == "" {
		p.config.UserFilter = "(uid=%s)"
	}
	if p.config.AttrMap == nil {
		p.config.AttrMap = map[string]string{
			"id":    "uid",
			"name":  "cn",
			"email": "mail",
		}
	}
	return nil
}

func (p *LDAPProvider) Authenticate(username, password string) (*OAuth2User, error) {
	conn, err := p.dial()
	if err != nil {
		return nil, fmt.Errorf("ldap dial: %w", err)
	}
	defer conn.Close()

	bindDN := p.config.BindDN
	bindPass := p.config.BindPassword
	if bindDN == "" {
		bindDN = fmt.Sprintf(p.config.UserFilter, username)
		if strings.Contains(p.config.UserFilter, "%s") {
			bindDN = fmt.Sprintf("uid=%s,%s", username, p.config.BaseDN)
		}
	}

	if err := conn.bind(bindDN, password); err != nil {
		return nil, fmt.Errorf("ldap bind: %w", err)
	}

	userDN := bindDN
	if p.config.BindDN != "" && bindPass != "" {
		userDN = fmt.Sprintf("uid=%s,%s", username, p.config.BaseDN)
		if err := conn.bind(userDN, password); err != nil {
			return nil, fmt.Errorf("ldap user bind: %w", err)
		}
	}

	filter := strings.ReplaceAll(p.config.UserFilter, "%s", username)
	attrs, err := conn.search(userDN, filter, p.config.BaseDN)
	if err != nil {
		return nil, fmt.Errorf("ldap search: %w", err)
	}

	user := &OAuth2User{ID: username, Name: username}
	for k, v := range p.config.AttrMap {
		if val, ok := attrs[strings.ToLower(v)]; ok {
			switch k {
			case "id":
				user.ID = val
			case "name":
				user.Name = val
			case "email":
				user.Email = val
			}
		}
	}

	return user, nil
}

func (p *LDAPProvider) GetAuthorizationURL(state string) string {
	return fmt.Sprintf("ldap://%s:%d", p.config.Host, p.config.Port)
}

func (p *LDAPProvider) ExchangeCode(code string) (*OAuth2Token, error) {
	return nil, fmt.Errorf("LDAP does not support code exchange; use Authenticate instead")
}

func (p *LDAPProvider) GetUser(token *OAuth2Token) (*OAuth2User, error) {
	return nil, fmt.Errorf("LDAP does not support GetUser; use Authenticate instead")
}

func (p *LDAPProvider) dial() (*ldapConn, error) {
	addr := net.JoinHostPort(p.config.Host, strconv.Itoa(p.config.Port))
	var conn net.Conn
	var err error

	var dialer net.Dialer
	ctx := context.Background()
	if p.config.Port == 636 {
		conn, err = (&tls.Dialer{NetDialer: &dialer, Config: &tls.Config{InsecureSkipVerify: false}}).DialContext(ctx, "tcp", addr)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return nil, err
	}

	return &ldapConn{conn: conn}, nil
}

func (c *ldapConn) Close() error {
	return c.conn.Close()
}

func (c *ldapConn) bind(dn, password string) error {
	// Simple LDAP BIND request (ASN.1 BER encoded)
	// LDAPMessage := SEQUENCE { msgID INTEGER, protocolOp CHOICE { bindRequest BindRequest } }
	// BindRequest := APPLICATION 0 SEQUENCE { version INTEGER, name OCTET STRING, authentication OCTET STRING }

	msgID := []byte{0x02, 0x01, 0x01}

	var dnBytes []byte
	if dn != "" {
		dnBytes = []byte(dn)
	}

	var passBytes []byte
	if password != "" {
		passBytes = []byte(password)
	}

	version := []byte{0x02, 0x01, 0x03} // LDAPv3
	name := encodeOctetString(dnBytes)
	auth := []byte{0x80, encLen(len(passBytes))}
	auth = append(auth, passBytes...)

	bindRequest := append(version, name...)
	bindRequest = append(bindRequest, auth...)
	bindRequest = append([]byte{0x60, encLen(len(bindRequest))}, bindRequest...)

	ldapMsg := append(msgID, bindRequest...)
	ldapMsg = append([]byte{0x30, encLen(len(ldapMsg))}, ldapMsg...)

	if _, err := c.conn.Write(ldapMsg); err != nil {
		return fmt.Errorf("write bind: %w", err)
	}

	resp := make([]byte, 1024)
	n, err := c.conn.Read(resp)
	if err != nil {
		return fmt.Errorf("read bind response: %w", err)
	}

	if n < 2 || resp[n-2] != 0x0a {
		return fmt.Errorf("ldap bind failed: invalid response")
	}

	resultCode := resp[n-1]
	if resultCode != 0 {
		codes := map[byte]string{
			0:  "success",
			1:  "operationsError",
			2:  "protocolError",
			3:  "timeLimitExceeded",
			4:  "sizeLimitExceeded",
			49: "invalidCredentials",
			50: "insufficientAccessRights",
		}
		msg, ok := codes[resultCode]
		if !ok {
			msg = fmt.Sprintf("unknown error code %d", resultCode)
		}
		return fmt.Errorf("ldap bind result: %s", msg)
	}

	return nil
}

func (c *ldapConn) search(_, filter, baseDN string) (map[string]string, error) {
	// Simple LDAP SEARCH request
	msgID := []byte{0x02, 0x01, 0x02}

	baseObject := encodeOctetString([]byte(baseDN))
	scope := []byte{0x0a, 0x01, 0x02} // wholeSubtree
	derefAliases := []byte{0x0a, 0x01, 0x00}
	sizeLimit := []byte{0x02, 0x01, 0x00}
	timeLimit := []byte{0x02, 0x01, 0x00}
	typesOnly := []byte{0x01, 0x01, 0x00}
	filterEnc := []byte{0xa7, encLen(len(filter) + 2)}
	filterEnc = append(filterEnc, []byte{0x04, encLen(len(filter))}...)
	filterEnc = append(filterEnc, []byte(filter)...)

	searchRequest := baseObject
	searchRequest = append(searchRequest, scope...)
	searchRequest = append(searchRequest, derefAliases...)
	searchRequest = append(searchRequest, sizeLimit...)
	searchRequest = append(searchRequest, timeLimit...)
	searchRequest = append(searchRequest, typesOnly...)
	searchRequest = append(searchRequest, filterEnc...)
	searchRequest = append([]byte{0x63, encLen(len(searchRequest))}, searchRequest...)

	ldapMsg := append(msgID, searchRequest...)
	ldapMsg = append([]byte{0x30, encLen(len(ldapMsg))}, ldapMsg...)

	if _, err := c.conn.Write(ldapMsg); err != nil {
		return nil, fmt.Errorf("write search: %w", err)
	}

	resp := make([]byte, 8192)
	n, err := c.conn.Read(resp)
	if err != nil {
		return nil, fmt.Errorf("read search: %w", err)
	}

	return parseLDAPResult(resp[:n]), nil
}

func encLen(n int) byte {
	if n > 255 {
		return 255
	}
	return byte(n)
}

func encodeOctetString(data []byte) []byte {
	return append([]byte{0x04, encLen(len(data))}, data...)
}

func parseLDAPResult(data []byte) map[string]string {
	result := make(map[string]string)

	_, entries := parseLDAPSequence(data)
	for _, entry := range entries {
		if len(entry) > 0 && entry[0] == 0x64 {
			_, parts := parseLDAPSequence(entry)
			for i := 1; i < len(parts); i++ {
				_, fields := parseLDAPSequence(parts[i])
				if len(fields) >= 2 {
					attrName := string(parseLDAPString(fields[0]))
					_, vals := parseLDAPSet(fields[1])
					if len(vals) > 0 {
						result[strings.ToLower(attrName)] = string(vals[0][2:])
					}
				}
			}
		}
	}

	return result
}

func parseLDAPSequence(data []byte) (int, [][]byte) {
	var items [][]byte
	pos := 0
	for pos < len(data) {
		if pos+2 > len(data) {
			break
		}
		tag := data[pos]
		length := int(data[pos+1])
		pos += 2
		if pos+length > len(data) {
			break
		}
		if tag == 0x30 || tag == 0x64 || tag == 0x31 {
			items = append(items, data[pos:pos+length])
		}
		pos += length
	}
	return pos, items
}

func parseLDAPString(data []byte) []byte {
	if len(data) < 2 {
		return nil
	}
	length := int(data[1])
	if len(data) < 2+length {
		return nil
	}
	return data[2 : 2+length]
}

func parseLDAPSet(data []byte) (int, [][]byte) {
	return parseLDAPSequence(data)
}
