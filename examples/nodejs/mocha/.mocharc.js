module.exports = {
  globalSetup: function () {
    process.env["PYROSCOPE_SAMPLING_DURATION"] = 1000;
    const Pyroscope = require("@pyroscope/nodejs");
    Pyroscope.init({ serverAddress: "_", appName: "example-mocha" });
    Pyroscope.start();
  },
  globalTeardown: function () {
    const Pyroscope = require("@pyroscope/nodejs");
    Pyroscope.stop();
  },
};
