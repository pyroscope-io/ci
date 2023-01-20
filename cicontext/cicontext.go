package cicontext

import "errors"

type BuildType string

const (
	PullRequest BuildType = "pull_request"
	Push        BuildType = "push"
	// TODO(eh-am): other types such as 'schedule'
)

type CIContext struct {
	Repo       string
	Owner      string
	CommitSHA  string
	BranchName string
	//	BuildType  string
}

var (
	ErrWrongCIProvider = errors.New("wrong ci provider")
)

type CIProvider interface {
	IsProvider() bool
	InferContext() (CIContext, error)
}

var providers []CIProvider = []CIProvider{
	&GithubActionsDetector{},
}

func detectCIProvider() (CIProvider, error) {
	// Find the first correct provider
	for _, p := range providers {
		if p.IsProvider() {
			return p, nil
		}
	}

	return nil, errors.New("could not infer correct ci provider, it may mean that it is not supported")
}

func Detect() (CIContext, error) {
	provider, err := detectCIProvider()
	if err != nil {
		return CIContext{}, err
	}

	ciCtx, err := provider.InferContext()
	if err != nil {
		return CIContext{}, err
	}

	return ciCtx, nil
}
