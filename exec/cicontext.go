package exec

import "github.com/pyroscope-io/ci/cicontext"

// Detect tries to detect the CIContext automatically
// If it cannot, it falls back to the values available in ExecCfg
// Note that it may return an invalid CIContext
func DetectContext(cfg ExecCfg) cicontext.CIContext {
	ciCtx, _ := cicontext.Detect()

	if cfg.Branch != "" {
		ciCtx.BranchName = cfg.Branch
	}

	if cfg.CommitSHA != "" {
		ciCtx.CommitSHA = cfg.CommitSHA
	}

	// TODO(eh-am): more fields
	return ciCtx
}
