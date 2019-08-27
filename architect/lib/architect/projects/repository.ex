defmodule Architect.Projects.Repository do
  @moduledoc """
  A process for interacting with a git repository

  On init this will immediately succeed, but trigger the `handle_continue` message as the first message handler.
  handle_continue will verify the repository in a new temporary directory.

  When this process is killed this directory should be automatically removed
  """

  defstruct [:dir, :address, :private_key, :verified, :fetched, :vcli]

  use Git.Repository
  require Logger
  alias Architect.VCLI
  alias Architect.Projects.Blueprint
  alias Git.{Branch, Commit}

  # Client

  # def start_link({process_name, address, private_key, known_hosts}) when is_binary(address) do
  #   Logger.debug("Starting repository process for #{address}")

  #   GenServer.start_link(__MODULE__, {address, private_key}, name: name, timeout: 10_000)
  # end

  @doc ~S"""
  Get the blueprints for a commit or branch
  """
  def list_blueprints(repository, selector) do
    GenServer.call(repository, {:list_blueprints, selector})
  end

  @doc ~S"""
  Get the project configuration for a repository from the default branch
  """
  def project_configuration(repository) do
    GenServer.call(repository, {:project_config})
  end

  @doc ~S"""
  Get the construction plan for a blueprint on a commit sha
  """
  def plan_blueprint(repository, branch_name, commit, blueprint_name) do
    GenServer.call(repository, {:plan_blueprint, branch_name, commit, blueprint_name})
  end

  # Server

  @impl true
  def handle_call(_, _from, %__MODULE__{verified: verified} = state) when verified != true do
    Logger.warn("Cannot perform action on un-verified repository")
    {:reply, :error, state}
  end

  @impl true
  def handle_call(_, _from, %__MODULE__{fetched: fetched} = state) when fetched != true do
    Logger.warn("Cannot perform action on un-fetched repository")
    {:reply, :error, state}
  end

  # @impl true
  # def handle_call(
  #       :fetch,
  #       _from,
  #       %__MODULE__{address: address, private_key: private_key, dir: dir} = state
  #     ) do
  #   case Git.Repository.Remote.fetch(address, private_key, nil, dir) do
  #     :ok ->
  #       {:reply, :ok, state}

  #     :error ->
  #       {:reply, :error, state}
  #   end
  # end

  @impl true
  def handle_call(
        {:list_blueprints, {:branch, branch}},
        _from,
        %__MODULE__{dir: dir, vcli: vcli} = state
      ) do
    {_out, 0} = System.cmd("git", ["checkout", branch], cd: dir)

    blueprints =
      VCLI.list(dir, vcli)
      |> Map.get("blueprints")
      |> Blueprint.parse()

    {:reply, blueprints, state}
  end

  @impl true
  def handle_call(
        {:project_config},
        _from,
        %__MODULE{dir: dir, vcli: vcli} = state
      ) do
    default_branch = Branch.default(dir)

    {_out, 0} = System.cmd("git", ["checkout", default_branch.name], cd: dir)

    project_config = VCLI.project_config(dir, vcli)

    {:reply, project_config, state}
  end

  @impl true
  def handle_call(
        {:plan_blueprint, branch_name, commit, blueprint_name},
        _from,
        %__MODULE{dir: dir, vcli: vcli} = state
      ) do
    {_out, 0} = System.cmd("git", ["checkout", commit], cd: dir)
    {_out, 0} = System.cmd("git", ["clean", "-fd"], cd: dir)

    blueprint_plan = VCLI.plan_blueprint(dir, vcli, branch_name, blueprint_name)

    {:reply, blueprint_plan, state}
  end
end
