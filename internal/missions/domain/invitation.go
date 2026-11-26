package domain

import "time"

// InvitationWindow is the confirmation window for team roster invitations.
// Once it lapses, a still-pending invitation resolves to EXPIRED.
const InvitationWindow = 7 * 24 * time.Hour

// InvitationState is the consent state of a team roster member.
type InvitationState string

const (
	// InvitationPending awaits the invitee's answer.
	InvitationPending InvitationState = "PENDING"
	// InvitationConfirmed means the invitee opted in.
	InvitationConfirmed InvitationState = "CONFIRMED"
	// InvitationDeclined means the invitee refused; they leave the roster.
	InvitationDeclined InvitationState = "DECLINED"
	// InvitationExpired means the window lapsed without an answer.
	InvitationExpired InvitationState = "EXPIRED"
)

// invitationTracksPending reports whether an invitation state is resolvable.
func invitationTracksPending(s InvitationState) bool { return s == InvitationPending }

// InvitationState returns the invitation state for a roster member, or empty
// when the member carries no invitation (solo contracts, pre-invitation data).
func (c *MissionContract) InvitationState(playerID string) InvitationState {
	return c.invitations[playerID]
}

// InvitedAt returns when the invitations were issued.
func (c *MissionContract) InvitedAt() time.Time { return c.invitedAt }

// Invitations returns a defensive copy of the invitation states by member.
func (c *MissionContract) Invitations() map[string]InvitationState {
	copied := make(map[string]InvitationState, len(c.invitations))
	for k, v := range c.invitations {
		copied[k] = v
	}
	return copied
}

// RestoreInvitationState rehydrates persisted invitation state. Infrastructure
// adapters call this after NewMissionContractFromState; it is a no-op for
// contracts that never carried invitations.
func (c *MissionContract) RestoreInvitationState(invitations map[string]InvitationState, invitedAt time.Time) {
	if len(invitations) == 0 {
		return
	}
	copied := make(map[string]InvitationState, len(invitations))
	for k, v := range invitations {
		copied[k] = v
	}
	c.invitations = copied
	c.invitedAt = invitedAt
	c.updateRosterFrozen()
}

// ConfirmInvitation resolves the caller's pending invitation as confirmed.
func (c *MissionContract) ConfirmInvitation(identity CallerIdentity, now time.Time) error {
	return c.resolveInvitation(identity.PlayerID(), InvitationConfirmed, now)
}

// DeclineInvitation resolves the caller's pending invitation as declined and
// removes them from the roster, releasing their reserved slot.
func (c *MissionContract) DeclineInvitation(identity CallerIdentity, now time.Time) error {
	return c.resolveInvitation(identity.PlayerID(), InvitationDeclined, now)
}

// ResolveExpirations expires every pending invitation past the window; expired
// members leave the roster. Invoked lazily by application commands.
func (c *MissionContract) ResolveExpirations(now time.Time) error {
	if c.invitedAt.IsZero() {
		return nil
	}
	for _, member := range c.roster {
		if c.invitations[member] == InvitationPending && now.Sub(c.invitedAt) >= InvitationWindow {
			c.invitations[member] = InvitationExpired
			c.removeFromRoster(member)
		}
	}
	c.updateRosterFrozen()
	return nil
}

func (c *MissionContract) resolveInvitation(playerID string, target InvitationState, now time.Time) error {
	if err := c.ResolveExpirations(now); err != nil {
		return err
	}
	if _, invited := c.invitations[playerID]; !invited {
		if c.IncludesPlayer(playerID) {
			return ErrInvitationNotPending
		}
		return ErrNotInvited
	}
	if !invitationTracksPending(c.invitations[playerID]) {
		return ErrInvitationNotPending
	}
	c.invitations[playerID] = target
	if target == InvitationDeclined {
		c.removeFromRoster(playerID)
	}
	c.updateRosterFrozen()
	return nil
}

// removeFromRoster drops a member whose invitation resolved negatively.
func (c *MissionContract) removeFromRoster(playerID string) {
	kept := c.roster[:0]
	for _, p := range c.roster {
		if p != playerID {
			kept = append(kept, p)
		}
	}
	c.roster = kept
}

// updateRosterFrozen freezes the roster once no invitation remains pending.
// Contracts without invitations freeze at creation.
func (c *MissionContract) updateRosterFrozen() {
	for _, state := range c.invitations {
		if invitationTracksPending(state) {
			c.rosterFrozen = false
			return
		}
	}
	c.rosterFrozen = true
}

// newPendingInvitations builds the invitation map for a team acceptance:
// the requester (roster[0]) is confirmed by accepting; everyone else starts
// pending. Solo rosters get no invitations.
func newPendingInvitations(roster []string, at time.Time) (map[string]InvitationState, time.Time) {
	if len(roster) <= 1 {
		return nil, time.Time{}
	}
	invitations := make(map[string]InvitationState, len(roster))
	for i, member := range roster {
		if i == 0 {
			invitations[member] = InvitationConfirmed
			continue
		}
		invitations[member] = InvitationPending
	}
	return invitations, at
}
