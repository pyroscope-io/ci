const { fibonacci } = require("./fib");
const { expect } = require("chai");

describe("fibonacci", () => {
  it("works", function () {
    this.timeout(0);

    expect(fibonacci(42)).to.equal(267914296);
  });
});
