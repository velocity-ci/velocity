use Mix.Config

# We don't run a server during test. If one is required,
# you can enable the server option below.
config :velocity, VelocityWeb.Endpoint,
  http: [port: 4001],
  server: false

# Print only warnings and errors during test
config :logger, level: :warn

# Configure your database
config :velocity, Velocity.Repo,
  adapter: Ecto.Adapters.Postgres,
  username: System.get_env("DATABASE_USERNAME") || "velocity_test",
  password: System.get_env("DATABASE_PASSWORD") || "velocity_test",
  database: System.get_env("DATABASE_NAME") || "velocity_test",
  hostname: System.get_env("DATABASE_HOSTNAME") || "localhost",
  port: System.get_env("DATABASE_PORT") || 5432,
  pool: Ecto.Adapters.SQL.Sandbox

  # Use less encryption rounds in test environment
config :comeonin, :bcrypt_log_rounds, 4