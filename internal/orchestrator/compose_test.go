package orchestrator

import (
	"strings"
	"testing"
)

func TestComposeScriptInline(t *testing.T) {
	script := composeScript(composeScriptParams{
		Stack:      "orbita-ab12cd34",
		InlineYAML: "services:\n  web:\n    image: nginx:alpine\n",
		Override:   buildComposeOverride("web", "orbita-org-acme", "orbita-ab12cd34", DeployConfig{}),
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
		Override:    buildComposeOverride("api", "orbita-org-acme", "orbita-ab12cd34", DeployConfig{}),
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
		Override:   buildComposeOverride("web", "orbita-org-acme", "orbita-ab12cd34", DeployConfig{}),
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

// Regression: app env vars must be injected into the *running* services, not
// just made available for interpolating the compose file at deploy time.
func TestComposeScriptInjectsEnvIntoServices(t *testing.T) {
	script := composeScript(composeScriptParams{
		Stack:      "orbita-ab12cd34",
		InlineYAML: "services:\n  web:\n    image: nginx\n  worker:\n    image: nginx\n",
		Override:   buildComposeOverride("web", "orbita-org-acme", "orbita-ab12cd34", DeployConfig{}),
		AppEnv:     renderEnvFile(map[string]string{"DATABASE_URL": "postgres://x"}),
	})

	for _, want := range []string{
		"/orbita-app.env",
		"echo 'services:' > /orbita-env.yml",
		// Applied to every service, not just web — workers need env too.
		"for s in $SERVICES; do",
		"    env_file:",
		"-c /orbita-env.yml",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("script missing %q\n---\n%s", want, script)
		}
	}
}

// With no env vars there should be no env file and no extra -c flag.
func TestComposeScriptOmitsEnvWhenEmpty(t *testing.T) {
	script := composeScript(composeScriptParams{
		Stack:      "orbita-ab12cd34",
		InlineYAML: "services:\n  web:\n    image: nginx\n",
		Override:   buildComposeOverride("web", "orbita-org-acme", "orbita-ab12cd34", DeployConfig{}),
	})
	if strings.Contains(script, "/orbita-env.yml") {
		t.Errorf("should not reference an env override when there are no vars\n---\n%s", script)
	}
}

// Verified against docker stack deploy: env_file values are literal to end of
// line. Adding quotes puts them *inside* the value, so nothing may be quoted or
// escaped.
func TestRenderEnvFileWritesValuesRaw(t *testing.T) {
	got := renderEnvFile(map[string]string{
		"URL":     "postgres://u:p@db/x",
		"QUOTED":  `he said "hi"`,
		"SPACES":  "hello world",
		"DOLLAR":  "$NOT_INTERPOLATED",
		"HASH":    "abc#def",
		"bad-key": "skipped",
		"MULTI":   "line1\nline2",
	})

	for _, want := range []string{
		"URL=postgres://u:p@db/x\n",
		"QUOTED=he said \"hi\"\n",
		"SPACES=hello world\n",
		"DOLLAR=$NOT_INTERPOLATED\n",
		"HASH=abc#def\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("env file missing %q\n---\n%s", want, got)
		}
	}
	if strings.Contains(got, `="`) {
		t.Errorf("values must not be wrapped in quotes\n---\n%s", got)
	}
	if strings.Contains(got, "bad-key") {
		t.Errorf("invalid env name should be skipped\n---\n%s", got)
	}
	// A multi-line value would corrupt every following variable.
	if strings.Contains(got, "MULTI") {
		t.Errorf("multi-line value should be skipped\n---\n%s", got)
	}
}

func TestComposeOverrideAppliesResourceLimits(t *testing.T) {
	got := buildComposeOverride("web", "orbita-org-acme", "orbita-ab12cd34",
		DeployConfig{MemoryMB: 512, CPUShares: 1500})

	for _, want := range []string{
		"    deploy:",
		"      resources:",
		"        limits:",
		"          memory: 512M",
		`          cpus: "1.50"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("override missing %q\n---\n%s", want, got)
		}
	}

	// No limits configured → no deploy block at all.
	none := buildComposeOverride("web", "orbita-org-acme", "orbita-ab12cd34", DeployConfig{})
	if strings.Contains(none, "deploy:") {
		t.Errorf("unset limits should not emit a deploy block\n---\n%s", none)
	}
}

// The override is what makes Traefik able to reach the stack: the web service
// joins the org network under the alias Traefik already routes to.
func TestComposeOverrideAliasesWebService(t *testing.T) {
	got := buildComposeOverride("api", "orbita-org-acme", "orbita-ab12cd34", DeployConfig{})
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
