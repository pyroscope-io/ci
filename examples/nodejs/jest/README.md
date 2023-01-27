# nodejs jest example

```
pyroscope-ci exec -- yarn test
```

# quirks
## `numWorkers=1`
As of now, it's necessary to run with a single worker (either via the `--runInBand` flag or the `--numWorkers=1`).
The reason is that there's no reliable way to run setup code per worker (see [issue in the Jest repo](https://github.com/facebook/jest/issues/8708)).
New approaches are being investigated.

## Hanging after test execution
The following message may be shown
```
Jest did not exit one second after the test run has completed.

This usually means that there are asynchronous operations that weren't stopped in your tests. Consider running Jest with `--detectOpenHandles` to troubleshoot this issue.
```

Which happens when the CPU profiler is still performing a 'round', which takes 10 seconds.

If you want to immediately quit, use the `--forceExit` flag, but be aware of its downsides.

## Disabling when running locally
An environment variable may be set in ci only, so that the tests are not slow when running locally.

```js
if (process.env.CI) {
  Pyroscope.start();
}
```
