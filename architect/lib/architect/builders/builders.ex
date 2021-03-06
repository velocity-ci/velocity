defmodule Architect.Builders do
  alias Architect.Builders.{Scheduler, Presence}

  use Supervisor

  require Logger

  ### Client

  def start_link(_opts \\ []) do
    Logger.debug("Starting #{Atom.to_string(__MODULE__)}")
    Supervisor.start_link(__MODULE__, :ok, name: __MODULE__)
  end

  def track(socket), do: Presence.track(socket)

  def update_status(socket, status), do: Presence.update_status(socket, status)

  def list(), do: Presence.list()

  def history(), do: Scheduler.history()

  ### Server

  def init(:ok) do
    children = [
      Scheduler,
      Presence
    ]

    state = Supervisor.init(children, strategy: :one_for_one)

    Logger.info("Running #{Atom.to_string(__MODULE__)}")

    state
  end
end
