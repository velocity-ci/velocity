defmodule Architect.Secretaries do
  alias Architect.Secretaries.{Presence, Secretary}
  alias Phoenix.{Socket, PubSub}

  use Supervisor

  require Logger

  ### Client

  def start_link(_opts \\ []) do
    Logger.debug("Starting #{Atom.to_string(__MODULE__)}")
    Supervisor.start_link(__MODULE__, :ok, name: __MODULE__)
  end

  def track(socket), do: Presence.track(socket)

  def list(), do: Presence.list()

  def history(), do: Scheduler.history()

  def get_commits(), do: Secretary.get_commits()

  ### Server

  def init(:ok) do
    children = [
      Presence,
      Secretary,
    ]

    state = Supervisor.init(children, strategy: :one_for_one)

    Logger.info("Running #{Atom.to_string(__MODULE__)}")

    state
  end

end
