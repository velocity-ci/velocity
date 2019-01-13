defmodule Architect.Secretaries.Secretary do
  use GenServer
  require Logger

  alias Phoenix.{PubSub, Socket}
  alias Architect.Secretaries.Presence

  defstruct [
    :history
  ]

  # Client

  def start_link(_opts) do
    Logger.debug("Starting process for #{Atom.to_string(__MODULE__)}")

    GenServer.start_link(__MODULE__, %__MODULE__{history: []}, name: __MODULE__)
  end

  def get_commits() do
    GenServer.call(__MODULE__, :get_commits)
  end

  # Server

  def init(state) do
    Logger.info("Running #{Atom.to_string(__MODULE__)}")

    PubSub.subscribe(Architect.PubSub, Presence.topic())

    {:ok, state}
  end

  def handle_info(
    %Socket.Broadcast{event: "presence_diff"} = broadcast,
    %__MODULE__{history: history} = state
  ) do
  Logger.debug(inspect(Presence.list()))
  Logger.debug(inspect(broadcast))

  {:noreply, %{state | history: [broadcast.payload | history]}}
  end

  def handle_call(:get_commits, _from, state) do
    # TODO: Should have a registry for secretaries holding what repositories are on each one

    secretary_sockets = Map.values(Presence.list())
    secretary_socket = List.first(secretary_sockets)
    IO.inspect(secretary_socket)

    %{metas: [metas]} = secretary_socket

    send(metas.socket, :get_commits)

    {:noreply, state, :hibernate}
  end
end
