defmodule Architect.Builders.Scheduler do
  @moduledoc """
  This manages and schedules Builds on Builders. Builder state is held in memory. We should put this onto ERTS/Mnesia.
  """
  use GenServer
  require Logger

  defstruct(builders: MapSet.new())

  #
  # Client API
  #

  def register_builder(builder) do
    {:ok, builder}
  end

  #
  # Server
  #

  def init(_opts) do
    state = %Architect.Builders.Scheduler{}

    Logger.info("-> started builder scheduler")

    # Poll for queued builds
    Process.send_after(Architect.Builders.Scheduler, :poll_builds, 10000)

    {:ok, state}
  end

  @doc """
  Internal(register_builder): Registers a builder in memory (ERTS in the future?)
  """
  @spec handle_call(tuple, String.t(), Architect.Builders.Scheduler) :: tuple
  def handle_call({:register_builder, builder}, _from, state) do
    IO.inspect(state.builders)

    {:reply, "", state, :hibernate}
  end

  @doc """
  Internal(poll_builds): Checks for any queued builds and attempts to schedule them onto a builder.
  """
  def handle_info(:poll_builds, _from, state) do
    Logger.info("checking for available builders")
    Logger.info("checking for waiting builds")
  end
end
