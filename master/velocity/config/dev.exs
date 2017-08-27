use Mix.Config

# For development, we disable any cache and enable
# debugging and code reloading.
#
# The watchers configuration can be used to run external
# watchers to your application. For example, we use it
# with brunch.io to recompile .js and .css sources.
config :velocity, VelocityWeb.Endpoint,
  http: [port: {:system, "PORT"}],
  debug_errors: true,
  code_reloader: true,
  check_origin: false,
  watchers: []

# Watch static and templates for browser reloading.
config :velocity, VelocityWeb.Endpoint,
  live_reload: [
    patterns: [
      ~r{priv/static/.*(js|css|png|jpeg|jpg|gif|svg)$},
      ~r{priv/gettext/.*(po)$},
      ~r{lib/velocity_web/views/.*(ex)$},
      ~r{lib/velocity_web/templates/.*(eex)$}
    ]
  ]

# Do not include metadata nor timestamps in development logs
config :logger, :console, format: "[$level] $message\n"

# Set a higher stacktrace during development. Avoid configuring such
# in production as building large stacktraces may be expensive.
config :phoenix, :stacktrace_depth, 20

# Configure your database
config :velocity, Velocity.Repo,
  adapter: Ecto.Adapters.Postgres,
  username: System.get_env("DATABASE_USERNAME") || "velocity_dev",
  password: System.get_env("DATABASE_PASSWORD") || "velocity_dev",
  database: System.get_env("DATABASE_NAME") || "velocity_dev",
  hostname: System.get_env("DATABASE_HOSTNAME") || "localhost",
  port: System.get_env("DATABASE_PORT") || 5432,
  pool_size: 10
