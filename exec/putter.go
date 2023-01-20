package exec

import (
	"context"
	"sync"

	"github.com/pyroscope-io/pyroscope/pkg/storage"
	"github.com/pyroscope-io/pyroscope/pkg/storage/metadata"
	"github.com/pyroscope-io/pyroscope/pkg/structs/flamebearer"
)

type Putter struct {
	sync.Mutex
	buffer FlamebearerMap
	// TODO: mutex
}

type FlamebearerMap map[string]*storage.PutInput

// NewPutter creates a Putter that stores Put data into memory
// Which can then be retrieved via GetPutItems
func NewPutter() *Putter {
	buffer := make(FlamebearerMap)

	return &Putter{
		buffer: buffer,
	}
}

func (p *Putter) Put(ctx context.Context, pi *storage.PutInput) error {
	p.Lock()
	if val, ok := p.buffer[pi.Key.AppName()]; ok {
		// This writes to val
		val.Val.Merge(pi.Val)
	} else {
		p.buffer[pi.Key.AppName()] = pi
	}
	p.Unlock()

	return nil
}

func (p *Putter) GetPutItems() map[string]flamebearer.FlamebearerProfile {
	dst := make(map[string]flamebearer.FlamebearerProfile, len(p.buffer))
	for i, v := range p.buffer {
		dst[i] = p.putInputToProfile(v)
	}

	return dst
}

func (p *Putter) putInputToProfile(pi *storage.PutInput) flamebearer.FlamebearerProfile {
	return flamebearer.NewProfile(flamebearer.ProfileConfig{
		Name:     pi.Key.AppName(),
		MaxNodes: -1,
		Tree:     pi.Val,
		Metadata: metadata.Metadata{
			SpyName:         pi.SpyName,
			SampleRate:      pi.SampleRate,
			Units:           pi.Units,
			AggregationType: pi.AggregationType,
		},
	})
}
