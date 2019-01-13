use Mix.Config

# We don't run a server during test. If one is required,
# you can enable the server option below.
config :architect, ArchitectWeb.Endpoint,
  http: [port: 4002],
  server: false

# Print only warnings and errors during test
config :logger, level: :warn

# Configure your database
config :architect, Architect.Repo,
  username: System.get_env("DB_USERNAME") || "velocity",
  password: System.get_env("DB_PASSWORD") || "velocity",
  database: System.get_env("DB_DBNAME") || "velocity_test",
  hostname: System.get_env("DB_HOSTNAME") || "localhost",
  port: String.to_integer(System.get_env("DB_PORT") || "5432"),
  pool: Ecto.Adapters.SQL.Sandbox
