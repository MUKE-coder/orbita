package service

import "testing"

func TestEnvVarNameForDB(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"mydb", "MYDB_URL"},
		{"rental-db", "RENTAL_DB_URL"},
		{"my.app db2", "MY_APP_DB2_URL"},
		{"POSTGRES", "POSTGRES_URL"},
	}
	for _, tt := range tests {
		if got := envVarNameForDB(tt.name); got != tt.want {
			t.Errorf("envVarNameForDB(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
