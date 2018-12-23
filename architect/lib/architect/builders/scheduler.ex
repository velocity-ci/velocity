defmodule Architect.Builders.Scheduler do
  @moduledoc """
  This manages and schedules Builds on Builders. Builder state is held in memory. We should put this onto ERTS/Mnesia.
  """
  use GenServer
  require Logger

  @poll_timeout 5000

  @enforce_keys [
    :name,
    :registry,
    :supervisor
  ]

  defstruct [
    :name,
    :registry,
    :supervisor
  ]

  #
  # Client API
  #
  def start_link(%__MODULE__{name: name} = state) do
    Logger.debug("Starting process for scheduler #{inspect(state)}")

    GenServer.start_link(__MODULE__, state, name: name)
  end

  #
  # Server
  #

  def init(%__MODULE__{name: name} = state) do
    Logger.info("Running #{Atom.to_string(__MODULE__)}")

    # Poll for queued builds
    Process.send_after(name, :poll_builds, @poll_timeout)

    {:ok, state}
  end

  def handle_info(:poll_builds, %__MODULE__{name: name} = state) do
    Logger.debug("checking for available builders")
    builders = Supervisor.which_children(Architect.Builders.Supervisor)
    # {:undefined, #PID<0.601.0>, :worker, [Architect.Builders.Builder]}
    Enum.each(builders, fn x ->
      {_undefined, pid, _something_else, _another_thing} = x
      state = Architect.Builders.Builder.get_state(pid)
      IO.inspect(state)
    end)

    Logger.debug("checking for waiting builds")

    Process.send_after(name, :poll_builds, @poll_timeout)

    {:noreply, state}
  end
end
