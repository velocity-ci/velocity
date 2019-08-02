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

  #
  # Server
  #
  @impl
  def init(state) do
    Logger.info("Running #{Atom.to_string(__MODULE__)}")

    builds = Architect.Builds.list_running_builds()

    Enum.each(builds, fn b ->
      changeset = Architect.Builds.Build.changeset(b, %{status: "waiting"})
      {:ok, _b} = Architect.Repo.update(changeset)
      Logger.info("Reset build:#{b.id} to 'waiting'")
    end)

    PubSub.subscribe(Architect.PubSub, Presence.topic())

    Process.send_after(Architect.Builders.Scheduler, :poll_builds, @poll_timeout)

    {:ok, state}
  end

  @impl
  def handle_call(:history, _from, state) do
    {:reply, state.history, state}
  end

  @impl
  def handle_info(:poll_builds, state) do
    Logger.debug("checking for waiting builds")
    builds = Architect.Builds.list_waiting_builds()

    Enum.each(builds, fn b ->
      Logger.debug("attempting to schedule build:#{b.id}")
      builders = Presence.list()

      Enum.each(builders, fn {id, %{metas: [metas]}} ->
        Logger.debug("builder #{id} (#{inspect(metas.socket)}) is #{metas.status}")
        send(metas.socket, b)
        changeset = Architect.Builds.Build.changeset(b, %{status: "running"})
        {:ok, _b} = Architect.Repo.update(changeset)
      end)
    end)

    Process.send_after(Architect.Builders.Scheduler, :poll_builds, @poll_timeout)
    {:noreply, state}
  end

  @impl
  def handle_info(
        %Socket.Broadcast{event: "presence_diff"} = broadcast,
        %__MODULE__{history: history} = state
      ) do
    Logger.debug(inspect(Presence.list()))
    Logger.debug(inspect(broadcast))

    {:noreply, %{state | history: [broadcast.payload | history]}}
  end
end
