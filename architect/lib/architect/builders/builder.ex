defmodule Architect.Builders.Builder do
  use GenServer
  require Logger

  @moduledoc """

  """

  @enforce_keys [:id]

  defstruct [
    :id,
    :token,
    :state,
    :created_at,
    :updated_at
  ]

  #### Client

  def start_link({%__MODULE__{} = state, name}) do
    Logger.debug("Starting process for builder #{inspect(state)}")

    GenServer.start_link(__MODULE__, state, name: name)
  end

  def echo(builder_pid) do
    GenServer.call(builder_pid, :echo)
  end

  def connect(builder_pid) do
    GenServer.call(builder_pid, {:update, :state, state_ready()})
  end

  def get_token(builder_pid) do
    GenServer.call(builder_pid, {:info, :token})
  end

  def get_state(builder_pid) do
    GenServer.call(builder_pid, {:info, :state})
  end

  ### Server

  def init(%__MODULE__{} = state) do
    Logger.info("Running process for builder #{inspect(state)}")

    {:ok, state}
  end

  def handle_call(:state, _from, state) do
    {:reply, state, state}
  end

  def handle_call({:info, field}, _from, state) do
    {:reply, Map.get(state, field), state}
  end

  def handle_call({:info, field}, _from, state) do
    {:reply, Map.get(state, field), state}
  end

  def handle_call({:update, field, value}, _from, state) do
    state = Map.put(state, field, value)
    {:reply, state, state}
  end

  def handle_call(:echo, _from, state) do
    Logger.debug("Echo!")
    {:reply, :echo, state}
  end

  def handle_cast(:echo, _from, state) do
    Logger.debug("Echo!")
    {:reply, state}
  end

  def state_ready, do: "ready"
  def state_busy, do: "busy"
  def state_error, do: "error"
  def state_disconnected, do: "disconnected"
end
