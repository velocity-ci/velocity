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

  def start_link(%__MODULE__{} = state) do
    Logger.debug("Starting process for builder #{inspect(state)}")

    GenServer.start_link(__MODULE__, state)
  end

  ### Server

  def init(%__MODULE__{} = state) do
    Logger.info("Running process for builder #{inspect(state)}")

    {:ok, state}
  end

  def state_ready, do: "ready"
  def state_busy, do: "busy"
  def state_error, do: "error"
  def state_disconnected, do: "disconnected"
end
