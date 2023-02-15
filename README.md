# pyroscope-ci
Profile your tests.

For more info, see the documentation [http://pyroscope.io/docs/ci](http://pyroscope.io/docs/ci)

# e2e tests
We use `testscript` + `docker` for testing

## Running
```bash
make test-e2e
```

## Adding tests

1. Create a `Dockerfile`
2. Add a test (see `e2e_test.go` for reference)
3. Write a `testscript.txtar`
