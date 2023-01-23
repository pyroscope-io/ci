package exec

import (
	"context"

	"github.com/pyroscope-io/pyroscope/pkg/ingestion"
	"github.com/pyroscope-io/pyroscope/pkg/storage"
	"github.com/pyroscope-io/pyroscope/pkg/structs/flamebearer"
)

type Ingester struct {
	putter          *Putter
	metricsExporter storage.MetricsExporter
}

func NewIngester() *Ingester {
	putter := NewPutter()
	return &Ingester{
		putter:          putter,
		metricsExporter: noopMetricsExporter{},
	}
}

func (i *Ingester) Ingest(ctx context.Context, in *ingestion.IngestInput) error {
	return in.Profile.Parse(ctx, i.putter, i.metricsExporter, in.Metadata)
}

func (i *Ingester) GetIngestedItems() map[string]flamebearer.FlamebearerProfile {
	return i.putter.GetPutItems()
}

type noopMetricsExporter struct{}

func (noopMetricsExporter) Evaluate(*storage.PutInput) (storage.SampleObserver, bool) {
	return nil, false
}
