const Pyroscope = require("@pyroscope/nodejs");

module.exports = function () {
  Pyroscope.stop();
};
