## naeos auth sso

SSO/SAML/LDAP provider management

### Synopsis

Configure and manage SSO identity providers (OIDC, SAML 2.0, LDAP).

Example:
  naeos auth sso configure oidc --name azure --issuer https://login.microsoftonline.com/... --client-id ... --client-secret ...
  naeos auth sso configure saml --name okta --sso-url https://... --entity-id ...
  naeos auth sso configure ldap --name corp-ad --host ad.example.com --base-dn dc=example,dc=com
  naeos auth sso list

```
naeos auth sso [flags]
```

### Options

```
  -h, --help   help for sso
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos auth](naeos_auth.md)	 - Authentication and authorization management
* [naeos auth sso configure](naeos_auth_sso_configure.md)	 - Configure an SSO provider
* [naeos auth sso list](naeos_auth_sso_list.md)	 - List configured SSO providers
* [naeos auth sso remove](naeos_auth_sso_remove.md)	 - Remove an SSO provider configuration

