const { fibonacci } = require("./fib");

describe("fibonacci", () => {
  it("works", () => {
    expect(fibonacci(42)).toBe(267914296);
  });
});
