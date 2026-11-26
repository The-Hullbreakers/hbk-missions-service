// Package membership provides the team-membership adapter. Party Manager is
// the authoritative source and is not yet integrated; the deny-all adapter
// rejects every declared teammate rather than trusting request input.
package membership

// DenyAllReader rejects every teammate membership check. It is the safe
// default until the real Party Manager integration lands: a permissive stub
// would let any caller consume another player's active-contract slot.
type DenyAllReader struct{}

// IsMember always reports not-a-member.
func (DenyAllReader) IsMember(string, string) (bool, error) {
	return false, nil
}
