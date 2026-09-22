package session_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
)

func TestJWTUserSession_HasRole_DirectRoles(t *testing.T) {
	s := session.NewJWTUserSession("", "client-id", nil)
	claims := map[string]any{"roles": []any{"admin", "player"}}

	if !s.HasRole(claims, "player") {
		t.Fatalf("expected role player to be found")
	}
	if s.HasRole(claims, "coach") {
		t.Fatalf("expected role coach to not be found")
	}
}

func TestJWTUserSession_HasRole_RealmAccess(t *testing.T) {
	s := session.NewJWTUserSession("", "client-id", nil)
	claims := map[string]any{
		"realm_access": map[string]any{"roles": []any{"admin"}},
	}

	if !s.HasRole(claims, "admin") {
		t.Fatalf("expected admin from realm_access")
	}
}

func TestJWTUserSession_HasRole_ResourceAccess(t *testing.T) {
	s := session.NewJWTUserSession("", "missions-service", nil)
	claims := map[string]any{
		"resource_access": map[string]any{
			"missions-service": map[string]any{"roles": []any{"operator"}},
		},
	}

	if !s.HasRole(claims, "operator") {
		t.Fatalf("expected operator from resource_access")
	}
}

func TestJWTUserSession_ValidateAudience_AcceptsAllowedAudience(t *testing.T) {
	s := session.NewJWTUserSession("", "client-id", []string{"aud-1"})
	claims := map[string]any{"aud": "aud-1"}

	if err := s.ValidateAudience(claims); err != nil {
		t.Fatalf("expected allowed audience to pass: %v", err)
	}
}

func TestJWTUserSession_ValidateAudience_RejectsDisallowedAudience(t *testing.T) {
	s := session.NewJWTUserSession("", "client-id", []string{"aud-1"})
	claims := map[string]any{"aud": "aud-2"}

	if err := s.ValidateAudience(claims); err == nil {
		t.Fatalf("expected disallowed audience to fail")
	}
}

func TestJWTUserSession_ValidateAudience_AcceptsNoAllowedAudience(t *testing.T) {
	s := session.NewJWTUserSession("", "client-id", nil)

	if err := s.ValidateAudience(map[string]any{}); err != nil {
		t.Fatalf("expected empty allowed audiences to pass")
	}
}

func TestJWTUserSession_ValidateAudience_HandlesInterfaceSlice(t *testing.T) {
	s := session.NewJWTUserSession("", "client-id", []string{"aud-1"})
	claims := map[string]any{"aud": []any{"aud-2", "aud-1"}}

	if err := s.ValidateAudience(claims); err != nil {
		t.Fatalf("expected interface slice to contain allowed audience: %v", err)
	}
}

func TestJWTUserSession_ValidateAudience_RejectsMissingAudience(t *testing.T) {
	s := session.NewJWTUserSession("", "client-id", []string{"aud-1"})

	if err := s.ValidateAudience(map[string]any{}); err == nil {
		t.Fatalf("expected missing audience to fail")
	}
}

func TestJWTUserSession_GetUserIDFromHeader_RequiresConfiguration(t *testing.T) {
	s := session.NewJWTUserSession("", "", nil)
	_, _, err := s.GetUserIDFromHeader("Bearer any")
	if err == nil {
		t.Fatalf("expected unconfigured session to reject token")
	}
}

func TestJWTUserSession_GetUserIDFromHeader_RequiresBearer(t *testing.T) {
	s := session.NewJWTUserSession("", "", nil)
	_, _, err := s.GetUserIDFromHeader("token")
	if err == nil {
		t.Fatalf("expected non-Bearer header to be rejected")
	}
}
