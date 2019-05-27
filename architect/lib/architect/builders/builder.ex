defmodule Architect.Builders.Builder do
  @moduledoc false

  require Logger

  alias Architect.Builders.StageScheduler
  alias Architect.Builds.Build

  use GenStage

  defmodule State do
    @attrs [:pid]
    defstruct @attrs
    @enforce_keys @attrs
  end

  #
  #  def child_spec(args) do
  #    %{
  #      id: :test_builder,
  #      start: {__MODULE__, :start_link, [args]},
  #      type: :worker
  #    }
  #  end

  @doc false
  def start_link(pid: pid) when is_pid(pid) do
    GenStage.start_link(__MODULE__, pid: pid)
  end

  def init(pid: pid) do
    {
      :consumer,
      %State{pid: pid},
      subscribe_to: [
        {StageScheduler, [max_demand: 1, min_demand: 0]}
      ]
    }
  end

  @doc false
  @impl true
  def handle_events([%Build{} = build], _from, %State{pid: pid} = state) do
    Logger.debug("Builder handling build #{inspect(build)}}")

    send(pid, build)

    receive do
      :completed ->
        Logger.debug("Build completed. Freeing up builder.")
    end

    {:noreply, [], state, :hibernate}
  end

  #  @doc false
  #  @impl true
  #  def init(args) do
  #    poller_name = {:via, Registry, {DlqHandler.PollerRegistry, {DlqHandler.Poller, task_queue}}}
  #
  #    {
  #      :consumer,
  #      %State{
  #        task_queue: task_queue,
  #        dead_letter_queue: dead_letter_queue,
  #        repo: repo,
  #        num: num
  #      },
  #      subscribe_to: [
  #        {poller_name, [max_demand: 1, min_demand: 0]}
  #      ]
  #    }
  #  end
end
