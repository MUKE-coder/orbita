package sshx

import "testing"

func TestKnownHostEntry(t *testing.T) {
	tests := []struct {
		name   string
		target Target
		want   string
	}{
		// Port 22 is implicit in known_hosts — stored as the bare host, which is
		// also what `ssh-keygen -R 1.2.3.4` expects.
		{"default port is bare", Target{Host: "213.136.89.197", Port: "22"}, "213.136.89.197"},
		{"custom port is bracketed", Target{Host: "213.136.89.197", Port: "2222"}, "[213.136.89.197]:2222"},
		{"hostname default port", Target{Host: "orbita.example.com", Port: "22"}, "orbita.example.com"},
		{"hostname custom port", Target{Host: "orbita.example.com", Port: "2222"}, "[orbita.example.com]:2222"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := KnownHostEntry(tt.target); got != tt.want {
				t.Errorf("KnownHostEntry(%+v) = %q, want %q", tt.target, got, tt.want)
			}
		})
	}
}

func TestForgetHostCommand(t *testing.T) {
	// The command we print must be the one a user can paste verbatim.
	got := ForgetHostCommand(Target{Host: "213.136.89.197", Port: "22"})
	want := "ssh-keygen -R 213.136.89.197"
	if got != want {
		t.Errorf("ForgetHostCommand() = %q, want %q", got, want)
	}

	got = ForgetHostCommand(Target{Host: "213.136.89.197", Port: "2222"})
	want = "ssh-keygen -R [213.136.89.197]:2222"
	if got != want {
		t.Errorf("ForgetHostCommand() non-default port = %q, want %q", got, want)
	}
}

// A client that never handshook has no key to compare — it must report an
// error rather than silently claiming the entry is absent/stale.
func TestKnownHostStateWithoutHostKey(t *testing.T) {
	c := &Client{target: Target{Host: "1.2.3.4", Port: "22"}}
	state, err := c.KnownHostState()
	if err == nil {
		t.Fatal("expected an error when no host key was captured, got nil")
	}
	if state != HostKeyAbsent {
		t.Errorf("state = %v, want HostKeyAbsent on error", state)
	}
}
