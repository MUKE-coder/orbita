package cmd

import "testing"

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"rental-manager": "rental-manager",
		"My App":         "my-app",
		"stoka":          "stoka",
		"Foo_Bar.Baz":    "foo-bar-baz",
		"  spaced  ":     "spaced",
		"UPPER":          "upper",
	}
	for in, want := range tests {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
