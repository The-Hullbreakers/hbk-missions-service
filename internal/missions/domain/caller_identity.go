package domain

// CallerIdentity is the authenticated caller resolved from the session.
// Request bodies never provide identity; this value object is the only source
// of player, role, and team information for business decisions.
type CallerIdentity struct {
	playerID string
	roles    []string
	teamIDs  []string
}

// NewCallerIdentity validates and builds a caller identity.
func NewCallerIdentity(playerID string, roles, teamIDs []string) (CallerIdentity, error) {
	if playerID == "" {
		return CallerIdentity{}, ErrCallerIdentityInvalid
	}
	return CallerIdentity{
		playerID: playerID,
		roles:    append([]string(nil), roles...),
		teamIDs:  append([]string(nil), teamIDs...),
	}, nil
}

// NewSystemIdentity builds an identity for internal actors (scheduler).
func NewSystemIdentity() CallerIdentity {
	return CallerIdentity{playerID: "system", roles: []string{"system"}}
}

// PlayerID returns the caller player.
func (i CallerIdentity) PlayerID() string { return i.playerID }

// Roles returns a copy of the caller roles.
func (i CallerIdentity) Roles() []string { return append([]string(nil), i.roles...) }

// HasRole reports whether the caller holds the given role.
func (i CallerIdentity) HasRole(role string) bool {
	for _, r := range i.roles {
		if r == role {
			return true
		}
	}
	return false
}

// TeamIDs returns a defensive copy of the caller team memberships.
func (i CallerIdentity) TeamIDs() []string { return append([]string(nil), i.teamIDs...) }

// IsTeamMember reports whether the caller belongs to the given team.
func (i CallerIdentity) IsTeamMember(teamID string) bool {
	for _, id := range i.teamIDs {
		if id == teamID {
			return true
		}
	}
	return false
}
