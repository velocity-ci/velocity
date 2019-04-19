defmodule Architect.Builds.ETSStore do
  use GenServer

  require Logger

  #
  # Client API
  #
  def start_link(_opts \\ []) do
    Logger.debug("Starting process for #{Atom.to_string(__MODULE__)}")

    GenServer.start_link(__MODULE__, %{}, name: __MODULE__)
  end

  def put_stream_line(stream_id, line_no, payload), do: GenServer.call(__MODULE__, {:put_stream_line, stream_id, line_no, payload})

  def put_step_update(step_id, payload), do: GenServer.call(__MODULE__, {:put_step_update, step_id, payload})

  def put_task_update(task_id, payload), do: GenServer.call(__MODULE__, {:put_task_update, task_id, payload})

  #
  # Server
  #

  @impl
  def init(state) do
    Logger.info("Running #{Atom.to_string(__MODULE__)}")
    :ets.new(:running_tasks, [:set, :protected, :named_table])
    :ets.new(:running_steps, [:set, :protected, :named_table])
    :ets.new(:running_streams, [:set, :protected, :named_table])
    {:ok, state}
  end

  @impl
  def handle_call({:put_stream_line, stream_id, line_no, payload}, _from, state) do
    res = :ets.insert(:running_streams, {"#{stream_id}:#{line_no}", payload})

    {:reply, res, state}
  end

  @impl
  def handle_call({:put_step_update, step_id, payload}, _from, state) do
    res = :ets.insert(:running_steps, {step_id, payload})

    {:reply, res, state}
  end

  @impl
  def handle_call({:put_task_update, task_id, payload}, _from, state) do
    res = :ets.insert(:running_tasks, {task_id, payload})

    {:reply, res, state}
  end

end
