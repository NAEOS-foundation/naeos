package auth

// Standard resource constants for RBAC.
const (
	ResourceSpec     = "spec"
	ResourcePipeline = "pipeline"
	ResourceArtifact = "artifact"
	ResourceProfile  = "profile"
	ResourcePlugin   = "plugin"
	ResourceCloud    = "cloud"
	ResourceAI       = "ai"
	ResourceConfig   = "config"
	ResourceAdmin    = "admin"
	ResourceAudit    = "audit"
	ResourceUser     = "user"
)

// Standard action constants for RBAC.
const (
	ActionRead   = "read"
	ActionWrite  = "write"
	ActionDelete = "delete"
	ActionAdmin  = "admin"
)

// RoutePermission maps a URL path to the required resource and action.
type RoutePermission struct {
	Resource string
	Action   string
}

// DefaultRolePermissions defines permissions for built-in roles.
var DefaultRolePermissions = map[string][]struct {
	Resource string
	Actions  []string
	Deny     bool
}{
	"admin": {
		{ResourceSpec, []string{ActionRead, ActionWrite, ActionDelete}, false},
		{ResourcePipeline, []string{ActionRead, ActionWrite, ActionDelete}, false},
		{ResourceArtifact, []string{ActionRead, ActionWrite, ActionDelete}, false},
		{ResourceProfile, []string{ActionRead, ActionWrite, ActionDelete}, false},
		{ResourcePlugin, []string{ActionRead, ActionWrite, ActionDelete}, false},
		{ResourceCloud, []string{ActionRead, ActionWrite, ActionDelete}, false},
		{ResourceAI, []string{ActionRead, ActionWrite}, false},
		{ResourceConfig, []string{ActionRead, ActionWrite}, false},
		{ResourceAdmin, []string{ActionAdmin}, false},
		{ResourceAudit, []string{ActionRead, ActionDelete}, false},
		{ResourceUser, []string{ActionRead, ActionWrite, ActionDelete}, false},
	},
	"developer": {
		{ResourceSpec, []string{ActionRead, ActionWrite}, false},
		{ResourcePipeline, []string{ActionRead, ActionWrite}, false},
		{ResourceArtifact, []string{ActionRead}, false},
		{ResourceProfile, []string{ActionRead}, false},
		{ResourcePlugin, []string{ActionRead}, false},
		{ResourceCloud, []string{ActionRead, ActionWrite}, false},
		{ResourceAI, []string{ActionRead, ActionWrite}, false},
	},
	"viewer": {
		{ResourceSpec, []string{ActionRead}, false},
		{ResourcePipeline, []string{ActionRead}, false},
		{ResourceArtifact, []string{ActionRead}, false},
		{ResourceProfile, []string{ActionRead}, false},
		{ResourcePlugin, []string{ActionRead}, false},
		{ResourceAI, []string{ActionRead}, false},
	},
}

var RoleTemplates = map[string][]struct {
	Resource string
	Actions  []string
	Deny     bool
}{
	"auditor": {
		{ResourceAudit, []string{ActionRead}, false},
		{ResourceSpec, []string{ActionRead}, false},
		{ResourcePipeline, []string{ActionRead}, false},
		{ResourceArtifact, []string{ActionRead}, false},
		{ResourceConfig, []string{ActionRead}, false},
		{ResourceAdmin, []string{ActionAdmin}, true},
		{ResourceUser, []string{ActionWrite, ActionDelete}, true},
	},
	"soc2_auditor": {
		{ResourceAudit, []string{ActionRead}, false},
		{ResourceSpec, []string{ActionRead}, false},
		{ResourcePipeline, []string{ActionRead}, false},
		{ResourceArtifact, []string{ActionRead}, false},
		{ResourceConfig, []string{ActionRead}, false},
		{ResourceAdmin, []string{ActionAdmin}, true},
		{ResourceUser, []string{ActionWrite, ActionDelete}, true},
		{ResourceCloud, []string{ActionRead}, false},
	},
	"gdpr_admin": {
		{ResourceUser, []string{ActionRead, ActionWrite, ActionDelete}, false},
		{ResourceAudit, []string{ActionRead}, false},
		{ResourceConfig, []string{ActionRead, ActionWrite}, false},
		{ResourceSpec, []string{ActionRead}, false},
		{ResourceAdmin, []string{ActionAdmin}, true},
		{ResourcePipeline, []string{ActionWrite, ActionDelete}, true},
	},
	"hipaa_admin": {
		{ResourceAudit, []string{ActionRead, ActionDelete}, false},
		{ResourceUser, []string{ActionRead, ActionWrite}, false},
		{ResourceConfig, []string{ActionRead, ActionWrite}, false},
		{ResourceSpec, []string{ActionRead}, false},
		{ResourceCloud, []string{ActionRead}, false},
		{ResourceAdmin, []string{ActionAdmin}, false},
		{ResourcePipeline, []string{ActionWrite, ActionDelete}, true},
	},
}

// SetupDefaultRoles creates the built-in roles (admin, developer, viewer)
// with their associated permissions on the given RBAC instance.
func SetupDefaultRoles(r *RBAC) {
	for roleName, perms := range DefaultRolePermissions {
		ra := make(map[string][]string, len(perms))
		deny := make(map[string][]string)
		for _, p := range perms {
			if p.Deny {
				deny[p.Resource] = p.Actions
			} else {
				ra[p.Resource] = p.Actions
			}
		}
		var denyMap map[string][]string
		if len(deny) > 0 {
			denyMap = deny
		}
		r.AddRole(&Role{Name: roleName, ResourceActions: ra, Deny: denyMap})
	}
}

func SetupRoleTemplate(r *RBAC, templateName, roleName string, parents []string) {
	perms, ok := RoleTemplates[templateName]
	if !ok {
		return
	}
	ra := make(map[string][]string, len(perms))
	deny := make(map[string][]string)
	for _, p := range perms {
		if p.Deny {
			deny[p.Resource] = p.Actions
		} else {
			ra[p.Resource] = p.Actions
		}
	}
	var denyMap map[string][]string
	if len(deny) > 0 {
		denyMap = deny
	}
	r.AddRole(&Role{Name: roleName, ResourceActions: ra, Deny: denyMap, Parents: parents})
}

func joinActions(actions []string) string {
	if len(actions) == 0 {
		return ""
	}
	out := actions[0]
	for _, a := range actions[1:] {
		out += "+" + a
	}
	return out
}
