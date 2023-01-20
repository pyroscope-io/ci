package exec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pyroscope-io/pyroscope/pkg/structs/flamebearer"
	"github.com/sirupsen/logrus"
)

type Exporter struct {
	logger    *logrus.Logger
	outputDir string
}

func NewExporter(logger *logrus.Logger, outputDir string) *Exporter {
	return &Exporter{
		logger:    logger,
		outputDir: outputDir,
	}
}

// Export exports each FlamebearerProfile into a JSON file
func (w *Exporter) Export(items map[string]flamebearer.FlamebearerProfile) error {
	// Only create the dir when there are items
	// So that we don't pollute with an empty dir
	if len(items) > 0 {
		if err := ensureDirExists(w.outputDir); err != nil {
			return err
		}
	}

	for k, v := range items {
		filename := filepath.Join(w.outputDir, fmt.Sprintf("%s.json", k))
		// Use JSON since that's what we are exporting
		if err := w.Write(filename, v); err != nil {
			return err
		}

		logrus.Infof("created %s", filename)
	}

	return nil
}

// Write writes a flamebearer in json format to its dataDir
func (w *Exporter) Write(path string, flame flamebearer.FlamebearerProfile) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(flame); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

func ensureDirExists(dir string) error {
	if dir == "" {
		return nil
	}

	if err := os.MkdirAll(dir, os.ModeDir|os.ModePerm); err != nil {
		return fmt.Errorf("could not create directory '%s': %w", dir, err)
	}

	return nil
}
