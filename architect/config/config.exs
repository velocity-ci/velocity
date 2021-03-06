# This file is responsible for configuring your application
# and its dependencies with the aid of the Mix.Config module.
#
# This configuration file is loaded before any dependency and
# is restricted to this project.

# General application configuration
use Mix.Config

config :elixir, ansi_enabled: true

config :architect,
  ecto_repos: [Architect.Repo],
  keyscan: [
    timeout: 7_000,
    log_errors: true
  ],
  vcli: [
    bin: "vcli",
    timeout: 7_000,
    log_errors: true
  ]

# Configures the endpoint
config :architect, ArchitectWeb.Endpoint,
  #  render_errors: [view: ArchitectWeb.V1.ErrorView, accepts: ~w(json)],
  pubsub: [name: Architect.PubSub, adapter: Phoenix.PubSub.PG2]

config :architect, Architect.Accounts,
  issuer: "velocity_architect",
  secret_key: "velocity_architect_dev_secret_key"

# Configures Elixir's Logger
config :logger, :console,
  format: "$time $metadata[$level] $message\n",
  metadata: [:request_id]

# Use Jason for JSON parsing in Phoenix
config :phoenix, :json_library, Jason

config :architect, Architect.Users.Guardian, issuer: "VelocityCI"

config :cors_plug,
  origin: ["http://localhost:3000", "http://localhost:3001", "http://velocity.local:3000", "http://velocity.local:3001"],
  max_age: 86400,
  methods: ["GET", "POST"]

# Import environment specific config. This must remain at the bottom
# of this file so it overrides the configuration defined above.
import_config "#{Mix.env()}.exs"
