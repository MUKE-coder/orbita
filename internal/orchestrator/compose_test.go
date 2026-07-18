package orchestrator

import (
	"strings"
	"testing"
)

func TestComposeScriptInline(t *testing.T) {
	script := composeScript(composeScriptParams{
		Stack:      "orbita-ab12cd34",
		InlineYAML: "services:\n  web:\n    image: nginx:alpine\n",
		Override:   buildComposeOverride("web", "orbita-org-acme", "orbita-ab12cd34"),
		DotEnv:     "FOO=bar\n",
	})

	// Inline stacks must not clone.
	if strings.Contains(script, "git clone") {
		t.Error("inline compose should not clone a repo")
	}
	for _, want := range []string{
		"set -e",
		"/work/docker-compose.yml",
		"/orbita-override.yml",
		"/work/.env",
		"docker stack deploy",
		"--prune",
		"'orbita-ab12cd34'",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("script missing %q\n---\n%s", want, script)
		}
	}
}

func TestComposeScriptFromGit(t *testing.T) {
	script := composeScript(composeScriptParams{
		Stack:       "orbita-ab12cd34",
		ComposePath: "deploy/stack.yml",
		CloneURL:    "https://x-access-token:secret@github.com/acme/app.git",
		Branch:      "main",
		Override:    buildComposeOverride("api", "orbita-org-acme", "orbita-ab12cd34"),
	})

	if !strings.Contains(script, "git clone --depth 1 --branch \"main\"") {
		t.Errorf("expected a shallow clone of the branch\n---\n%s", script)
	}
	// The compose path must be resolved under the clone dir, not the repo root.
	if !strings.Contains(script, "/work/deploy/stack.yml") {
		t.Errorf("expected compose path under /work\n---\n%s", script)
	}
	// No .env line when there are no env vars.
	if strings.Contains(script, "/work/.env") {
		t.Error("should not write .env when no env vars are set")
	}
}

// The build path must both build and then pin the resulting tags, because
// `docker stack deploy` ignores build: and rejects services with no image.
func TestComposeScriptPinsBuiltImages(t *testing.T) {
	script := composeScript(composeScriptParams{
		Stack:      "orbita-ab12cd34",
		InlineYAML: "services:\n  web:\n    build: .\n",
		Override:   buildComposeOverride("web", "orbita-org-acme", "orbita-ab12cd34"),
	})

	for _, want := range []string{
		"docker compose -p 'orbita-ab12cd34'",
		"config --services",
		"docker image inspect \"orbita-ab12cd34-$s:latest\"",
		"echo \"    image: orbita-ab12cd34-$s:latest\" >> /orbita-images.yml",
		"IMAGES_ARG='-c /orbita-images.yml'",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("script missing %q\n---\n%s", want, script)
		}
	}
}

// The override is what makes Traefik able to reach the stack: the web service
// joins the org network under the alias Traefik already routes to.
func TestComposeOverrideAliasesWebService(t *testing.T) {
	got := buildComposeOverride("api", "orbita-org-acme", "orbita-ab12cd34")
	for _, want := range []string{
		"  api:",
		"      default: {}",
		"      orbita-org-acme:",
		"          - orbita-ab12cd34",
		"    external: true",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("override missing %q\n---\n%s", want, got)
		}
	}
}
