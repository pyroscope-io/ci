const Pyroscope = require("@pyroscope/nodejs");

module.exports = function () {
  Pyroscope.init({
    appName: "pyroscope.tests",
    serverAddress: "noop",
  });

  Pyroscope.start();
};
