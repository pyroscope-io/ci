package install

import (
	"fmt"
	"strings"
)

// ProfileTypesToCode converts ProfileTypes flags into the name it's used in the agent code
// TODO: unify this with upstream
// https://github.com/pyroscope-io/client/blob/46ac3c0a285626dbf2ef17e017fae2e1b985ee12/pyroscope/types.go#L14-L26
func ProfileTypesToCode(p string) ([]string, error) {
	if p == "all" {
		return []string{
			`pyroscope.ProfileCPU`,
			`pyroscope.ProfileInuseObjects`,
			`pyroscope.ProfileAllocObjects`,
			`pyroscope.ProfileInuseSpace`,
			`pyroscope.ProfileAllocSpace`,
			`pyroscope.ProfileGoroutines`,
			`pyroscope.ProfileMutexCount`,
			`pyroscope.ProfileMutexDuration`,
			`pyroscope.ProfileBlockCount`,
			`pyroscope.ProfileBlockDuration`,
		}, nil
	}

	types := strings.Split(p, ",")
	chosen := make([]string, 0)

	for _, profileType := range types {
		switch profileType {
		case "all":
			return nil, fmt.Errorf("type 'all' must be set alone")
		case "cpu":
			chosen = append(chosen, `pyroscope.ProfileCPU`)
		case "inuse_objects":
			chosen = append(chosen, `pyroscope.ProfileInuseObjects`)
		case "alloc_objects":
			chosen = append(chosen, `pyroscope.ProfileAllocObjects`)
		case "inuse_space":
			chosen = append(chosen, `pyroscope.ProfileInuseSpace`)
		case "alloc_space":
			chosen = append(chosen, `pyroscope.ProfileAllocSpace`)
		case "goroutines":
			chosen = append(chosen, `pyroscope.ProfileGoroutines`)
		case "mutex_count":
			chosen = append(chosen, `pyroscope.ProfileMutexCount`)
		case "mutex_duration":
			chosen = append(chosen, `pyroscope.ProfileMutexDuration`)
		case "block_count":
			chosen = append(chosen, `pyroscope.ProfileBlockCount`)
		case "block_duration":
			chosen = append(chosen, `pyroscope.ProfileBlockDuration`)
		}
	}

	if len(chosen) <= 0 {
		return nil, fmt.Errorf("at least a single valid profileType must be set")
	}

	return chosen, nil
}
