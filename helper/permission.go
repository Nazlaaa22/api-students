package helper

type PermissionSet map[string]bool

func (p PermissionSet) Has(permission string) bool {
	return p[permission]
}

func PermissionsForRole(role string) PermissionSet {
	switch role {
	case "admin":
		return PermissionSet{
			"student:list":       true,
			"student:read:any":   true,
			"student:create":     true,
			"student:update:any": true,
			"student:delete":     true,
		}

	case "staff":
		return PermissionSet{
			"student:list":     true,
			"student:read:any": true,
			"student:create":   true,
		}

	case "user":
		return PermissionSet{}

	default:
		return PermissionSet{}
	}
}
