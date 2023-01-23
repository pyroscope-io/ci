package cicontext_test

import (
	"testing"

	"github.com/pyroscope-io/ci/cicontext"
)

func TestGithubActions(t *testing.T) {
	gh := cicontext.GithubActionsDetector{}

	// Setup env
	t.Setenv("GITHUB_ACTION", "foo")
	t.Setenv("GITHUB_SHA", "sha")
	t.Setenv("GITHUB_EVENT_NAME", "pull_request")
	t.Setenv("GITHUB_HEAD_REF", "ref")
	t.Setenv("GITHUB_REPOSITORY", "owner/repository")

	vars, err := gh.InferContext()
	if err != nil {
		t.Fatal(err)
	}

	if vars.BranchName != "ref" {
		t.Fatal("wrong branchName")
	}

	if vars.CommitSHA != "sha" {
		t.Fatal("wrong commitSHA")
	}

	if vars.Owner != "owner" {
		t.Fatal("wrong owner")
	}

	if vars.Repo != "repository" {
		t.Fatal("wrong repo")
	}
}
