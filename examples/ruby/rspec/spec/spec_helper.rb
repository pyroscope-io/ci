require "bundler/setup"
require "rspec"
require "pyroscope"

RSpec.configure do |config|
  # Enable flags like --only-failures and --next-failure
  config.example_status_persistence_file_path = ".rspec_status"

  # Disable RSpec exposing methods globally on `Module` and `main`
  config.disable_monkey_patching!

  config.expect_with :rspec do |c|
    c.syntax = :expect
  end
end

# Enable pyroscope
RSpec.configure do |config|
  config.before(:suite) do
    Pyroscope.configure do |config|
      config.server_address = ENV["PYROSCOPE_ADHOC_SERVER_ADDRESS"]
      config.application_name = "ruby.tests"
    end
  end
end

