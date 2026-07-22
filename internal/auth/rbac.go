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
}{
	"admin": {
		{ResourceSpec, []string{ActionRead, ActionWrite, ActionDelete}},
		{ResourcePipeline, []string{ActionRead, ActionWrite, ActionDelete}},
		{ResourceArtifact, []string{ActionRead, ActionWrite, ActionDelete}},
		{ResourceProfile, []string{ActionRead, ActionWrite, ActionDelete}},
		{ResourcePlugin, []string{ActionRead, ActionWrite, ActionDelete}},
		{ResourceCloud, []string{ActionRead, ActionWrite, ActionDelete}},
		{ResourceAI, []string{ActionRead, ActionWrite}},
		{ResourceConfig, []string{ActionRead, ActionWrite}},
		{ResourceAdmin, []string{ActionAdmin}},
		{ResourceAudit, []string{ActionRead, ActionDelete}},
		{ResourceUser, []string{ActionRead, ActionWrite, ActionDelete}},
	},
	"developer": {
		{ResourceSpec, []string{ActionRead, ActionWrite}},
		{ResourcePipeline, []string{ActionRead, ActionWrite}},
		{ResourceArtifact, []string{ActionRead}},
		{ResourceProfile, []string{ActionRead}},
		{ResourcePlugin, []string{ActionRead}},
		{ResourceCloud, []string{ActionRead, ActionWrite}},
		{ResourceAI, []string{ActionRead, ActionWrite}},
	},
	"viewer": {
		{ResourceSpec, []string{ActionRead}},
		{ResourcePipeline, []string{ActionRead}},
		{ResourceArtifact, []string{ActionRead}},
		{ResourceProfile, []string{ActionRead}},
		{ResourcePlugin, []string{ActionRead}},
		{ResourceAI, []string{ActionRead}},
	},
}

// SetupDefaultRoles creates the built-in roles (admin, developer, viewer)
// with their associated permissions on the given RBAC instance.
func SetupDefaultRoles(r *RBAC) {
	for roleName, perms := range DefaultRolePermissions {
		ra := make(map[string][]string, len(perms))
		for _, p := range perms {
			ra[p.Resource] = p.Actions
		}
		r.AddRole(&Role{Name: roleName, ResourceActions: ra})
	}
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
