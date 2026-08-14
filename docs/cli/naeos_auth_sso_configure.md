## naeos auth sso configure

Configure an SSO provider

### Synopsis

Configure an SSO identity provider for authentication.

Supported provider types: oidc, saml, ldap

OIDC example:
  naeos auth sso configure oidc --name azure --issuer https://login.microsoftonline.com/tenant/v2.0 --client-id xxx --client-secret yyy --redirect-url http://localhost:8080/callback

SAML example:
  naeos auth sso configure saml --name okta --sso-url https://dev-xxxx.okta.com/app/.../sso/saml --entity-id https://example.com --cert-file ./okta-cert.pem

LDAP example:
  naeos auth sso configure ldap --name corp-ad --host ad.example.com --port 389 --base-dn dc=example,dc=com --bind-dn cn=admin,dc=example,dc=com --bind-password secret

```
naeos auth sso configure [flags]
```

### Options

```
      --base-dn string         LDAP base DN
      --bind-dn string         LDAP bind DN
      --bind-password string   LDAP bind password
      --cert-file string       SAML x509 certificate file path
      --client-id string       OAuth2 client ID
      --client-secret string   OAuth2 client secret
      --entity-id string       SAML entity ID
  -h, --help                   help for configure
      --host string            LDAP host
      --issuer string          OIDC issuer URL
      --name string            provider name (required)
      --port int               LDAP port (636 for LDAPS) (default 389)
      --provider-type string   
      --redirect-url string    OAuth2 redirect URL (default "http://localhost:8080/callback")
      --scope stringArray      OAuth2 scopes
      --sso-url string         SAML SSO URL
      --user-filter string     LDAP user filter (default: (uid=%s))
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos auth sso](naeos_auth_sso.md)	 - SSO/SAML/LDAP provider management

