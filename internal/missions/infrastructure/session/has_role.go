package session

// HasRole checks whether the claims contain the requested role.
func (s *JWTUserSession) HasRole(claims map[string]any, role string) bool {
	if rolesClaim, ok := claims["roles"].([]interface{}); ok {
		for _, r := range rolesClaim {
			if roleStr, ok := r.(string); ok && roleStr == role {
				return true
			}
		}
	}

	if roleClaim, ok := claims["role"].(string); ok && roleClaim == role {
		return true
	}

	if tenantRoles, ok := claims["tenant_roles"].([]interface{}); ok {
		for _, r := range tenantRoles {
			if roleStr, ok := r.(string); ok && roleStr == role {
				return true
			}
		}
	}

	if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			for _, r := range roles {
				if roleStr, ok := r.(string); ok && roleStr == role {
					return true
				}
			}
		}
	}

	if resourceAccess, ok := claims["resource_access"].(map[string]interface{}); ok {
		if clientAccess, ok := resourceAccess[s.clientID].(map[string]interface{}); ok {
			if roles, ok := clientAccess["roles"].([]interface{}); ok {
				for _, r := range roles {
					if roleStr, ok := r.(string); ok && roleStr == role {
						return true
					}
				}
			}
		}
	}

	return false
}
