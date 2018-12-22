defmodule Architect.Builders do
  alias Architect.Builders.Builder
  use Supervisor
  require Logger

  @registry Architect.Builders.Registry
  @supervisor Architect.Builders.Supervisor

  ### Client

  def start_link(_opts \\ []) do
    Logger.debug("Starting #{Atom.to_string(__MODULE__)}")
    Supervisor.start_link(__MODULE__, :ok, name: __MODULE__)
  end

  def start_builder(%Builder{id: id} = builder) do
    case Registry.lookup(@registry, id) do
      [{_, found}] ->
        Logger.error(
          "Attempted to start builder #{inspect(builder)} but id already exists in registry"
        )

        :error

      [] ->
        {:ok, pid} = DynamicSupervisor.start_child(@supervisor, {Builder, builder})
        {:ok, _} = Registry.register(@registry, id, pid)

        :ok
    end
  end

  def get_builder_state(id) do
    Registry.dispatch(@registry, id, fn
      [found] -> true
      [] -> false
    end)
  end

  ### Server

  def init(:ok) do
    children = [
      {Registry, keys: :unique, name: @registry},
      {DynamicSupervisor, name: @supervisor, strategy: :one_for_one}
    ]

    state = Supervisor.init(children, strategy: :one_for_one)

    Logger.info("Running #{Atom.to_string(__MODULE__)}")

    state
  end
end
