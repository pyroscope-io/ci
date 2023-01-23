package cicontext

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-multierror"
)

type GithubActionsDetector struct{}

func (*GithubActionsDetector) IsProvider() bool {
	return os.Getenv("GITHUB_ACTION") != ""
}

// InferContext infers the CIContext using environment variables
// It assumes the correct provider was identified via IsProvider
// For more reference of what each environment variable represents,
// Check Github Action's documentation at https://docs.github.com/en/actions/learn-github-actions/environment-variables
func (g *GithubActionsDetector) InferContext() (CIContext, error) {
	var ciCtx CIContext
	var errs error

	ciCtx.CommitSHA = os.Getenv("GITHUB_SHA")
	if ciCtx.CommitSHA == "" {
		errs = multierror.Append(errs, fmt.Errorf("could not identify CommitSHA"))
	}

	ciCtx.BranchName = g.inferBranch()
	if ciCtx.BranchName == "" {
		errs = multierror.Append(errs, fmt.Errorf("could not identify the branch"))
	}

	fullRepo := os.Getenv("GITHUB_REPOSITORY")
	fullRepoSlice := strings.Split(fullRepo, "/")

	if len(fullRepoSlice) != 2 {
		errs = multierror.Append(errs, fmt.Errorf("could not identify the repository/owner"))
		return ciCtx, errs
	}

	ciCtx.Repo = fullRepoSlice[1]
	if ciCtx.Repo == "" {
		errs = multierror.Append(errs, fmt.Errorf("could not identify the repository"))
	}

	ciCtx.Owner = fullRepoSlice[0]
	if ciCtx.Owner == "" {
		errs = multierror.Append(errs, fmt.Errorf("could not identify the owner"))
	}

	return ciCtx, errs
}

func (*GithubActionsDetector) isPR() bool {
	return os.Getenv("GITHUB_EVENT_NAME") == "pull_request"
}

func (g *GithubActionsDetector) inferBranch() string {
	// Since in a PR, github performs a merge commit
	// The branch name is not representative of the actual pull request
	if g.isPR() {
		return os.Getenv("GITHUB_HEAD_REF")
	}

	return os.Getenv("GITHUB_REF_NAME")
}
