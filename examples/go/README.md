# Usage

First, install the profiler into each package with a test:
`pyro-ci go install --applicationName=myapp fib/`

Then execute the test command
`pyro-ci exec -- go test ./...`

You may want to use `-count=1` to [bypass test caching](https://go.dev/doc/go1.10#test).
