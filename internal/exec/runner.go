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
	handler  http.Handler
	logger   *logrus.Logger
	ingester *Ingester
}

func NewRunner(logger *logrus.Logger) *Runner {
	ingester := NewIngester()
	handler := server.NewIngestHandler(logger, ingester, func(*ingestion.IngestInput) {}, httputils.NewDefaultHelper(logger))

	return &Runner{
		ingester: ingester,
		handler:  handler,
		logger:   logger,
	}
}

// Run executes a command while running a server that ingests data at the /ingest endpoint
// And returns all data that was ingested
func (p *Runner) Run(args []string) (map[string]flamebearer.FlamebearerProfile, time.Duration, error) {
	var m map[string]flamebearer.FlamebearerProfile

	mux := http.NewServeMux()
	mux.Handle("/ingest", p.handler)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		var zeroDuration time.Duration
		return m, zeroDuration, err
	}
	p.logger.Debugf("Ingester listening to port %d", listener.Addr().(*net.TCPAddr).Port)

	done := make(chan error)
	go func() {
		done <- http.Serve(listener, mux)
	}()

	// start command
	c := make(chan os.Signal, 10)
	// Note that we don't specify which signals to be sent: any signal to be
	// relayed to the child process (including SIGINT and SIGTERM).
	signal.Notify(c)

	startingTime := time.Now().UTC()
	captureDuration := func() time.Duration {
		return time.Since(startingTime)
	}

	// TODO: dirty hack to allow pushing to a proxy
	// Use case is running in docker, where a proxy forwards to the host machine
	ingesterAddress := "http://localhost"
	if os.Getenv("PYROSCOPE_PROXY_ADDRESS") != "" {
		ingesterAddress = "http://" + os.Getenv("PYROSCOPE_PROXY_ADDRESS")
	}

	env := fmt.Sprintf("PYROSCOPE_ADHOC_SERVER_ADDRESS=%s:%d", ingesterAddress, listener.Addr().(*net.TCPAddr).Port)
	cmd := osExec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(), env)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		return p.ingester.GetIngestedItems(), captureDuration(), err
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
			return p.ingester.GetIngestedItems(), captureDuration(), err
		case <-ticker.C:
			if !process.Exists(cmd.Process.Pid) {
				logrus.Debug("child process exited")
				err := cmd.Wait()

				if exiterr, ok := err.(*osExec.ExitError); ok {
					err = fmt.Errorf("exit code %d: %v", exiterr.ExitCode(), err)
				}

				return p.ingester.GetIngestedItems(), captureDuration(), err
			}
		}
	}
}
