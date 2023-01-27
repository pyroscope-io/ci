const { fibonacci } = require("./fib");

describe("fibonacci", () => {
  it("works", () => {
    expect(fibonacci(49)).toBe(7778742049);
  });
});
