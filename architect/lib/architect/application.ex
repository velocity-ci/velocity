defmodule Architect.Application do
  # See https://hexdocs.pm/elixir/Application.html
  # for more information on OTP Applications
  @moduledoc false

  use Application
  import Supervisor.Spec

  def start(_type, _args) do
    # List all child processes to be supervised
    children = [
      # Start the Ecto repository
      Architect.Repo,
      # Start the endpoint when the application starts
      ArchitectWeb.Endpoint,
      # Starts a worker by calling: Architect.Worker.start_link(arg)
      # {Architect.Worker, arg},
      # {Absinthe.Subscription, [ArchitectWeb.Endpoint]}
      supervisor(Absinthe.Subscription, [ArchitectWeb.Endpoint])
    ]

    # See https://hexdocs.pm/elixir/Supervisor.html
    # for other strategies and supported options
    opts = [strategy: :one_for_one, name: Architect.Supervisor]
    Supervisor.start_link(children, opts)
  end

  # Tell Phoenix to update the endpoint configuration
  # whenever the application is updated.
  def config_change(changed, _new, removed) do
    ArchitectWeb.Endpoint.config_change(changed, removed)
    :ok
  end
end
