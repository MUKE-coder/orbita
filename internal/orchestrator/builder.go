package orchestrator

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
)

type BuildContext struct {
	RepoURL       string
	Branch        string
	RootDir       string
	BuildMethod   string // dockerfile, nixpacks, buildpack
	DockerfilePath string
	BuildArgs     map[string]string
	OrgSlug       string
	AppName       string
	CommitSHA     string
}

type BuildResult struct {
	ImageRef string
	Logs     string
}

func (o *Orchestrator) BuildFromDockerfile(ctx context.Context, bc BuildContext) (*BuildResult, error) {
	// TODO: real impl
	// 1. git clone --depth 1 --branch {branch} {repo_url} to temp dir
	// 2. Docker BuildKit API to build image
	// 3. Push to local registry: localhost:5000/orbita/{orgSlug}/{appName}:{commit-sha}
	// 4. Clean up temp dir

	imageRef := fmt.Sprintf("localhost:5000/orbita/%s/%s:%s", bc.OrgSlug, bc.AppName, bc.CommitSHA)
	log.Info().
		Str("repo", bc.RepoURL).
		Str("branch", bc.Branch).
		Str("method", bc.BuildMethod).
		Str("image", imageRef).
		Msg("Building from Dockerfile (stub)")

	return &BuildResult{
		ImageRef: imageRef,
		Logs:     fmt.Sprintf("Cloning %s@%s...\nBuilding with %s...\nPushing %s...\nDone.\n", bc.RepoURL, bc.Branch, bc.BuildMethod, imageRef),
	}, nil
}

func (o *Orchestrator) BuildWithNixpacks(ctx context.Context, bc BuildContext) (*BuildResult, error) {
	// TODO: real impl — run nixpacks CLI in Docker container
	imageRef := fmt.Sprintf("localhost:5000/orbita/%s/%s:%s", bc.OrgSlug, bc.AppName, bc.CommitSHA)
	log.Info().
		Str("repo", bc.RepoURL).
		Str("branch", bc.Branch).
		Str("image", imageRef).
		Msg("Building with Nixpacks (stub)")

	return &BuildResult{
		ImageRef: imageRef,
		Logs:     fmt.Sprintf("Detecting language...\nBuilding with Nixpacks...\nPushing %s...\nDone.\n", imageRef),
	}, nil
}
