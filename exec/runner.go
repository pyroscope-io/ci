package exec

import (
	"fmt"
	"net"
	"net/http"
	"os"
	osExec "os/exec"
	"os/signal"
	"time"

	"github.com/pyroscope-io/pyroscope/pkg/ingestion"
	"github.com/pyroscope-io/pyroscope/pkg/server"
	"github.com/pyroscope-io/pyroscope/pkg/server/httputils"
	"github.com/pyroscope-io/pyroscope/pkg/structs/flamebearer"
	"github.com/pyroscope-io/pyroscope/pkg/util/process"
	"github.com/sirupsen/logrus"
)

type Runner struct {
	//	args    []string
	handler  http.Handler
	logger   *logrus.Logger
	ingester *Ingester
}

func NewRunner(logger *logrus.Logger) (*Runner, error) {
	ingester := NewIngester()
	handler := server.NewIngestHandler(logger, ingester, func(*ingestion.IngestInput) {}, httputils.NewDefaultHelper(logger), true)

	return &Runner{
		ingester: ingester,
		handler:  handler,
		logger:   logger,
	}, nil
}

// Run executes a command while running a server that ingests data at the /ingest endpoint
// And returns all data that was ingested
func (p *Runner) Run(args []string) (map[string]flamebearer.FlamebearerProfile, time.Duration, error) {
	var m map[string]flamebearer.FlamebearerProfile
	var duration time.Duration

	http.Handle("/ingest", p.handler)
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return m, duration, err
	}
	p.logger.Debugf("Ingester listening to port %d", listener.Addr().(*net.TCPAddr).Port)

	done := make(chan error)
	go func() {
		done <- http.Serve(listener, nil)
	}()

	// start command
	c := make(chan os.Signal, 10)
	// Note that we don't specify which signals to be sent: any signal to be
	// relayed to the child process (including SIGINT and SIGTERM).
	signal.Notify(c)
	env := fmt.Sprintf("PYROSCOPE_ADHOC_SERVER_ADDRESS=http://localhost:%d", listener.Addr().(*net.TCPAddr).Port)
	startingTime := time.Now().UTC()
	cmd := osExec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), env)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		return m, duration, err
	}
	defer func() {
		signal.Stop(c)
		close(c)
	}()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case s := <-c:
			_ = process.SendSignal(cmd.Process, s)
		case err := <-done:
			return m, duration, err
		case <-ticker.C:
			if !process.Exists(cmd.Process.Pid) {
				logrus.Debug("child process exited")
				duration = time.Now().Sub(startingTime)

				return p.ingester.GetIngestedItems(), duration, cmd.Wait()
			}
		}
	}
}
