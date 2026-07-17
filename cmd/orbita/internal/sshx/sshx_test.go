package sshx

import "testing"

func TestParseTarget(t *testing.T) {
	tests := []struct {
		in               string
		user, host, port string
		wantErr          bool
	}{
		{"root@1.2.3.4", "root", "1.2.3.4", "22", false},
		{"deploy@example.com:2222", "deploy", "example.com", "2222", false},
		{"1.2.3.4", "root", "1.2.3.4", "22", false},
		{"example.com:22", "root", "example.com", "22", false},
		{"", "", "", "", true},
		{"user@", "", "", "", true},
	}
	for _, tt := range tests {
		got, err := ParseTarget(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseTarget(%q) err=%v wantErr=%v", tt.in, err, tt.wantErr)
			continue
		}
		if tt.wantErr {
			continue
		}
		if got.User != tt.user || got.Host != tt.host || got.Port != tt.port {
			t.Errorf("ParseTarget(%q) = %+v, want %s/%s/%s", tt.in, got, tt.user, tt.host, tt.port)
		}
	}
}
