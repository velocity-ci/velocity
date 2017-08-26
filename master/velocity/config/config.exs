# This file is responsible for configuring your application
# and its dependencies with the aid of the Mix.Config module.
#
# This configuration file is loaded before any dependency and
# is restricted to this project.
use Mix.Config

# General application configuration
config :velocity,
  ecto_repos: [Velocity.Repo]

# Configures the endpoint
config :velocity, VelocityWeb.Endpoint,
  url: [host: "localhost"],
  secret_key_base: "vK6DpekKScNXdDvTzOmkmhST0nJRqgAHPCg6qzunjszZNjm7bw4y8oGuR6JLnPVH",
  render_errors: [view: VelocityWeb.ErrorView, accepts: ~w(html json)],
  pubsub: [name: Velocity.PubSub,
           adapter: Phoenix.PubSub.PG2]

# Configures Elixir's Logger
config :logger, :console,
  format: "$time $metadata[$level] $message\n",
  metadata: [:request_id]

# Import environment specific config. This must remain at the bottom
# of this file so it overrides the configuration defined above.
import_config "#{Mix.env}.exs"
