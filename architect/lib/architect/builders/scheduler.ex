defmodule Architect.Builders.Scheduler do
  @moduledoc """
  This manages and schedules Tasks on Builders. Builder state is held in memory. We should put this onto ETS/Mnesia.
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
  def init(state) do
    Logger.info("Running #{Atom.to_string(__MODULE__)}")

    tasks = Architect.Builds.list_building_tasks()

    Enum.each(tasks, fn t ->
      changeset = Architect.Builds.Task.update_changeset(t, %{status: "ready"})
      {:ok, _b} = Architect.Repo.update(changeset)
      Logger.info("Reset task:#{t.id} to 'ready'")
    end)

    PubSub.subscribe(Architect.PubSub, Presence.topic())

    Process.send_after(Architect.Builders.Scheduler, :poll_builds, @poll_timeout)

    {:ok, state}
  end

  def handle_call(:history, _from, state) do
    {:reply, state.history, state}
  end

  defp schedule_task(task) do
    builders = Presence.list()

    Enum.each(builders, fn {id, %{metas: [metas]}} ->
      Logger.debug("builder #{id} (#{inspect(metas.socket)}) is #{metas.status}")
      Logger.debug("scheduling task: #{task.id}")

      if metas.status == :ready do
        stage =
          Ecto.assoc(task, :stage)
          |> Architect.Repo.one()

        build =
          Ecto.assoc(stage, :build)
          |> Architect.Repo.one()
          |> Architect.Repo.preload(:project)

        send(metas.socket, {build, task})
        changeset = Architect.Builds.Task.update_changeset(task, %{status: "building"})
        {:ok, _t} = Architect.Repo.update(changeset)
      end
    end)
  end

  def handle_info(:poll_builds, state) do
    Logger.debug("checking for ready tasks")
    tasks = Architect.Builds.list_ready_tasks()
    Logger.debug("found #{length(tasks)} tasks")

    Enum.each(tasks, fn task ->
      task
      |> schedule_task()
    end)

    Process.send_after(Architect.Builders.Scheduler, :poll_builds, @poll_timeout)
    {:noreply, state}
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
