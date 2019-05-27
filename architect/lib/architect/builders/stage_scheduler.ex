defmodule Architect.Builders.StageScheduler do
  @moduledoc false

  use GenStage
  require Logger
  alias DlqHandler.Message
  alias DlqHandler.Builds.Build

  @behaviour GenStage

  def child_spec(args) do
    %{
      id: __MODULE__,
      start: {__MODULE__, :start_link, args},
      type: :worker
    }
  end

  @doc false
  def start_link() do
    GenStage.start_link(__MODULE__, [], name: __MODULE__)
  end

  @doc false
  @impl true
  def init(_) do
    Process.send(self(), :get_waiting_builds, [])

    {:producer, :no_state, demand: :accumulate}
  end

  @impl true
  def handle_demand(new_demand, :no_state) do
    IO.puts("new demand #{inspect(new_demand)}")

    {:noreply, [], :no_state}
  end

  @impl true
  def handle_info(:get_waiting_builds, :no_state) do
    builds = Architect.Builds.list_waiting_builds()

    Process.send_after(self(), :get_waiting_builds, 30_000)

    {:noreply, builds, :no_state, :hibernate}
  end
end
