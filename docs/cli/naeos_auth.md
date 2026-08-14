## naeos auth

Authentication and authorization management

### Synopsis

Manage users, roles, API keys, and OAuth2 providers.

Example:
  naeos auth whoami --api-key <key>
  naeos auth create-user --name john --email john@example.com --role admin
  naeos auth create-key --user-id u1 --name my-api-key
  naeos auth list-users
  naeos auth list-roles

```
naeos auth [flags]
```

### Options

```
  -h, --help   help for auth
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos auth assign-role](naeos_auth_assign-role.md)	 - Assign a role to a user
* [naeos auth create-key](naeos_auth_create-key.md)	 - Create a new API key
* [naeos auth create-role](naeos_auth_create-role.md)	 - Create a custom RBAC role
* [naeos auth create-role-from-template](naeos_auth_create-role-from-template.md)	 - Create a role from a compliance template (auditor, soc2_auditor, gdpr_admin, hipaa_admin)
* [naeos auth create-user](naeos_auth_create-user.md)	 - Create a new user
* [naeos auth delete-role](naeos_auth_delete-role.md)	 - Delete an RBAC role
* [naeos auth list-role-templates](naeos_auth_list-role-templates.md)	 - List available role templates for compliance frameworks
* [naeos auth list-roles](naeos_auth_list-roles.md)	 - List all roles
* [naeos auth list-users](naeos_auth_list-users.md)	 - List all users
* [naeos auth login](naeos_auth_login.md)	 - Login via OAuth2 provider
* [naeos auth logout](naeos_auth_logout.md)	 - Logout current session
* [naeos auth sso](naeos_auth_sso.md)	 - SSO/SAML/LDAP provider management
* [naeos auth whoami](naeos_auth_whoami.md)	 - Show current authenticated identity

