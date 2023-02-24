require 'spec_helper'
require_relative '../fib'

RSpec.describe do
  it "calculates fib" do
    expect(fib(42)).to eq(267914296)
  end
end
