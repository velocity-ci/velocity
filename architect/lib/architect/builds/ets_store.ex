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

  def put_stream_line(stream_id, line_no, payload),
    do: GenServer.call(__MODULE__, {:put_stream_line, stream_id, line_no, payload})

  def put_step_update(step_id, payload),
    do: GenServer.call(__MODULE__, {:put_step_update, step_id, payload})

  def put_task_update(task_id, payload),
    do: GenServer.call(__MODULE__, {:put_task_update, task_id, payload})

  def get_stream_lines(stream_id),
    do: GenServer.call(__MODULE__, {:get_stream_lines, stream_id})

  #
  # Server
  #

  @impl
  def init(state) do
    Logger.info("Running #{Atom.to_string(__MODULE__)}")
    :ets.new(:build_tasks, [:set, :protected, :named_table])
    :ets.new(:build_steps, [:set, :protected, :named_table])
    :ets.new(:build_streams, [:bag, :protected, :named_table])
    {:ok, state}
  end

  @impl
  def handle_call({:put_stream_line, stream_id, line_no, payload}, _from, state) do
    res = :ets.insert(:build_streams, {"#{stream_id}", {line_no, Poison.encode!(payload)}})

    {:reply, res, state}
  end

  @impl
  def handle_call({:put_step_update, step_id, payload}, _from, state) do
    res = :ets.insert(:build_steps, {step_id, Poison.encode!(payload)})

    {:reply, res, state}
  end

  @impl
  def handle_call({:put_task_update, task_id, payload}, _from, state) do
    res = :ets.insert(:build_tasks, {task_id, Poison.encode!(payload)})

    {:reply, res, state}
  end

  @impl
  def handle_call({:get_stream_lines, stream_id}, _from, state) do
    res =
      :ets.lookup(:build_streams, stream_id)
      |> Enum.map(fn {stream_id, {line_no, payload}} ->
        Poison.decode!(payload)
      end)

    {:reply, res, state}
  end
end
