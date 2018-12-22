defmodule Architect.Builders do
  alias Architect.Builders.{Builder, Scheduler}
  use Supervisor
  use GenServer
  require Logger

  @registry Architect.Builders.Registry
  @supervisor Architect.Builders.Supervisor
  @scheduler Architect.Builders.Scheduler

  @registry_str Atom.to_string(@registry)

  ### Client

  def start_link(_opts \\ []) do
    Logger.debug("Starting #{Atom.to_string(__MODULE__)}")
    Supervisor.start_link(__MODULE__, :ok, name: __MODULE__)
  end

  def start_builder(%Builder{id: id} = builder) do
    case Registry.lookup(@registry, id) do
      [_] ->
        Logger.error(
          "Failed to register builder #{inspect(builder)} on #{@registry_str}; id already exists"
        )

        {:error, :already_exists}

      [] ->
        name = {:via, Registry, {@registry, id}}
        {:ok, pid} = DynamicSupervisor.start_child(@supervisor, {Builder, {builder, name}})

        Logger.debug("Registered builder #{inspect(builder)} on #{@registry_str}")

        {:ok, pid}
    end
  end

  def stop_builder(id) do
    case Registry.lookup(@registry, id) do
      [{pid, _}] ->
        DynamicSupervisor.terminate_child(@supervisor, pid)

      [] ->
        Logger.error(
          "Failed to stop builder #{inspect(id)} on #{@registry_str}; id does not exist"
        )

        {:error, :not_found}
    end
  end

  def call_builder(id, request) do
    case Registry.lookup(@registry, id) do
      [{pid, _}] ->
        try do
          GenServer.call(pid, request)
        catch
          kind, reason ->
            formatted = Exception.format(kind, reason, __STACKTRACE__)

            Logger.error(
              "Failed to call builder #{inspect(id)} on #{@registry_str} with #{formatted}"
            )
        end

      [] ->
        Logger.error(
          "Failed to call builder #{inspect(id)} on #{@registry_str}; id does not exist"
        )

        {:error, :not_found}
    end
  end

  ### Server

  def init(:ok) do
    children = [
      {Registry, keys: :unique, name: @registry},
      {DynamicSupervisor, name: @supervisor, strategy: :one_for_one},
      {Scheduler, %Scheduler{name: @scheduler, registry: @registry, supervisor: @supervisor}}
    ]

    state = Supervisor.init(children, strategy: :one_for_one)

    Logger.info("Running #{Atom.to_string(__MODULE__)}")

    state
  end
end
