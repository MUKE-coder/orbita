package models

import "testing"

func TestHasMinRole(t *testing.T) {
	tests := []struct {
		userRole, minRole string
		want              bool
	}{
		// hierarchy: viewer < developer < admin < owner
		{RoleViewer, RoleViewer, true},
		{RoleViewer, RoleDeveloper, false},
		{RoleViewer, RoleAdmin, false},
		{RoleViewer, RoleOwner, false},
		{RoleDeveloper, RoleViewer, true},
		{RoleDeveloper, RoleDeveloper, true},
		{RoleDeveloper, RoleAdmin, false},
		{RoleAdmin, RoleDeveloper, true},
		{RoleAdmin, RoleOwner, false},
		{RoleOwner, RoleOwner, true},
		{RoleOwner, RoleViewer, true},

		// unknown/empty roles must never pass any check
		{"", RoleViewer, false},
		{"superuser", RoleViewer, false},
		{"OWNER", RoleViewer, false}, // roles are lowercase; wrong case is unknown
	}

	for _, tt := range tests {
		if got := HasMinRole(tt.userRole, tt.minRole); got != tt.want {
			t.Errorf("HasMinRole(%q, %q) = %v, want %v", tt.userRole, tt.minRole, got, tt.want)
		}
	}
}
