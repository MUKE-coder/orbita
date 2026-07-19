package service

import "testing"

// The default posture is what most operators get, so it carries the most cases:
// exactly one public sign-up, then the door closes and invites are the way in.
func TestRegistrationAllowed(t *testing.T) {
	cases := []struct {
		name         string
		userCount    int64
		hasInvite    bool
		allowOpen    bool
		hardDisabled bool
		want         bool
	}{
		// Bootstrap
		{"first user on a fresh instance", 0, false, false, false, true},
		{"first user even when hard-disabled", 0, false, false, true, true},
		{"first user with open registration", 0, false, true, false, true},

		// Default posture: closed after the first account
		{"second user, no invite, defaults", 1, false, false, false, false},
		{"hundredth user, no invite, defaults", 99, false, false, false, false},

		// Invites are the way in
		{"invited user on a closed instance", 1, true, false, false, true},
		{"invited user, open instance", 1, true, true, false, true},

		// Hard disable beats everything except bootstrap
		{"invited user but hard-disabled", 1, true, false, true, false},
		{"open flag but hard-disabled", 1, false, true, true, false},

		// Explicitly opened
		{"open registration, no invite", 1, false, true, false, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := registrationAllowed(c.userCount, c.hasInvite, c.allowOpen, c.hardDisabled)
			if got != c.want {
				t.Errorf("registrationAllowed(count=%d, invite=%v, open=%v, disabled=%v) = %v, want %v",
					c.userCount, c.hasInvite, c.allowOpen, c.hardDisabled, got, c.want)
			}
		})
	}
}

// The two failure modes that matter, stated as their own assertions so a
// regression names the actual risk rather than a table row.
func TestRegistrationClosesAfterFirstUserByDefault(t *testing.T) {
	if !registrationAllowed(0, false, false, false) {
		t.Fatal("a fresh instance must accept its first sign-up, or it can never be set up")
	}
	if registrationAllowed(1, false, false, false) {
		t.Fatal("public sign-up must close once an account exists — otherwise anyone can register on a stranger's server")
	}
}

func TestInviteOpensClosedInstance(t *testing.T) {
	if registrationAllowed(1, false, false, false) {
		t.Fatal("precondition: instance should be closed")
	}
	if !registrationAllowed(1, true, false, false) {
		t.Fatal("an invited user must be able to create an account, or invites are useless on a closed instance")
	}
}
