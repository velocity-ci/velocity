defmodule Architect.Builders.Scheduler do
  @moduledoc """
  This manages and schedules Builds on Builders. Builder state is held in memory. We should put this onto ERTS/Mnesia.
  """
  alias Architect.Builders.Presence
  alias Phoenix.PubSub
  alias Phoenix.Socket

  use GenServer

  require Logger

  @poll_timeout 5000

  defstruct [
    :history
  ]

  #
  # Client API
  #
  def start_link(_opts) do
    Logger.debug("Starting process for #{Atom.to_string(__MODULE__)}")

    GenServer.start_link(__MODULE__, %__MODULE__{history: []}, name: __MODULE__)
  end

  def history() do
    GenServer.call(__MODULE__, :history)
  end

  def list(), do: Presence.list()

  def track(socket), do: Presence.track(socket)

  #
  # Server
  #

  def init(state) do
    Logger.info("Running #{Atom.to_string(__MODULE__)}")

    PubSub.subscribe(Architect.PubSub, Presence.topic())

    {:ok, state}
  end

  def handle_call(:history, _from, state) do
    {:reply, state.history, state}
  end

  def handle_info(
        %Socket.Broadcast{event: "presence_diff"} = broadcast,
        %__MODULE__{history: history} = state
      ) do
    Logger.debug(inspect(Presence.list()))
    Logger.debug(inspect(broadcast))

    {:noreply, %{state | history: [broadcast.payload | history]}}
  end
end
