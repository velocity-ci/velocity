defmodule Architect.Projects.Repository do
  @moduledoc """
  A process for interacting with a git repository

  On init this will immediately succeed, but trigger the `handle_continue` message as the first message handler.
  handle_continue will verify the repository in a new temporary directory.

  When this process is killed this directory should be automatically removed
  """

  defstruct [:dir, :address, :private_key, :verified, :fetched, :vcli]

  use GenServer
  require Logger
  alias Architect.VCLI
  alias Architect.Projects.Blueprint
  alias Git.{Branch, Commit}

  # Client

  def start_link({address, private_key, name}) when is_binary(address) do
    Logger.debug("Starting fresh repository process for #{address}")

    GenServer.start_link(__MODULE__, {address, private_key}, name: name, timeout: 10_000)
  end

  @doc ~S"""
  Check if verified
  """
  def verified?(repository), do: GenServer.call(repository, :verified)

  @doc ~S"""
  Get commit amount across all branches
  """
  def commit_count(repository), do: GenServer.call(repository, :commit_count)

  @doc ~S"""
  Get commit amount for branch
  """
  def commit_count(repository, branch), do: GenServer.call(repository, {:commit_count, branch})

  @doc ~S"""
  Run `git fetch` on the repository
  """
  def fetch(repository), do: GenServer.call(repository, :fetch)

  @doc ~S"""
  Get a single commit by its SHA value
  """
  def commit_by_sha(repository, sha), do: GenServer.call(repository, {:get_commit_by_sha, sha})

  @doc ~S"""
  Get a list of commits by branch
  """
  def list_commits(repository, branch), do: GenServer.call(repository, {:list_commits, branch})

  @doc ~S"""
  Get a list of branches
  """
  def list_branches(repository), do: GenServer.call(repository, :list_branches)

  @doc ~S"""
  Get a single branch
  """
  def get_branch(repository, branch), do: GenServer.call(repository, {:get_branch, branch})

  @doc ~S"""
  Get a list of branches for a commit SHA value
  """
  def list_branches_for_commit(repository, sha),
    do: GenServer.call(repository, {:list_branches_for_commit, sha})

  @doc ~S"""
  Get the default branch
  """
  def default_branch(repository), do: GenServer.call(repository, :default_branch)

  @doc ~S"""
  Ge the blueprints for a commit or branch
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
  def init({address, private_key}) when is_binary(address) do
    {:ok,
     %__MODULE__{
       address: address,
       private_key: private_key,
       verified: false,
       vcli: VCLI.init()
     }, {:continue, :verify}}
  end

  @impl true
  def handle_continue(:verify, %__MODULE__{address: address, private_key: private_key} = state)
      when is_binary(address) do
    Temp.track!()
    repo_dir = Temp.mkdir!(Slugger.slugify(address))

    with %Porcelain.Result{err: nil, out: _, status: 0} <-
           Porcelain.exec("git", ["init"], dir: repo_dir),
         %Porcelain.Result{err: nil, out: _, status: 0} <-
           Porcelain.exec("git", ["remote", "add", "origin", address], dir: repo_dir),
         :ok <- Git.Repository.Remote.verify(address, private_key, repo_dir) do
      send(self(), :fetch)

      {:noreply, %__MODULE__{state | dir: repo_dir, verified: true}}
    else
      %Porcelain.Result{err: err, out: out, status: _} ->
        Logger.error("Failed verifying #{address} in #{repo_dir}. #{err}: #{out}")
        {:noreply, %__MODULE__{state | verified: false}}

      :error ->
        Logger.error("Failed verifying #{address} in #{repo_dir}.")
        {:noreply, %__MODULE__{state | verified: false}}
    end
  end

  @impl true
  def handle_call(:verified, _from, %__MODULE__{verified: verified} = state) do
    {:reply, verified, state}
  end

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

  @impl true
  def handle_call(
        :fetch,
        _from,
        %__MODULE__{address: address, private_key: private_key, dir: dir} = state
      ) do
    case Git.Repository.Remote.fetch(address, private_key, dir) do
      :ok ->
        {:reply, :ok, state}

      :error ->
        {:reply, :error, state}
    end
  end

  @impl true
  def handle_call(:list_branches, _from, %__MODULE__{dir: dir} = state) do
    branches = Git.Branch.list(dir)

    {:reply, branches, state}
  end

  def handle_call({:get_branch, branch}, _from, %__MODULE__{dir: dir} = state) do
    branch =
      Git.Branch.list(dir)
      |> Enum.find(fn %Git.Branch{name: b} -> b == branch end)

    {:reply, branch, state}
  end

  @impl true
  def handle_call({:list_branches_for_commit, sha}, _from, %__MODULE__{dir: dir} = state) do
    branches = Git.Branch.list_for_commit_sha(dir, sha)

    {:reply, branches, state}
  end

  @impl true
  def handle_call({:list_commits, branch}, _from, %__MODULE{dir: dir} = state) do
    commits = Commit.list_for_ref(dir, branch)

    {:reply, commits, state}
  end

  @impl true
  def handle_call({:get_commit_by_sha, sha}, _from, %__MODULE{dir: dir} = state) do
    commit = Commit.get_by_sha(dir, sha)

    {:reply, commit, state}
  end

  @impl true
  def handle_call(:default_branch, _from, %__MODULE__{dir: dir} = state) do
    branch = Branch.default(dir)

    {:reply, branch, state}
  end

  @impl true
  def handle_call({:commit_count, branch}, _from, %__MODULE__{dir: dir} = state) do
    count = Commit.count_for_branch(dir, branch)

    {:reply, count, state}
  end

  @impl true
  def handle_call(:commit_count, _from, %__MODULE__{dir: dir} = state) do
    count = Commit.count(dir)

    {:reply, count, state}
  end

  @impl true
  def handle_call(
        {:list_blueprints, {:branch, branch}},
        _from,
        %__MODULE{dir: dir, vcli: vcli} = state
      ) do
    %Porcelain.Result{err: nil, out: _, status: 0} =
      Porcelain.exec("git", ["checkout", branch], dir: dir)

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

    %Porcelain.Result{err: nil, out: _, status: 0} =
      Porcelain.exec("git", ["checkout", default_branch.name], dir: dir)

    project_config = VCLI.project_config(dir, vcli)

    {:reply, project_config, state}
  end

  @impl true
  def handle_call(
        {:plan_blueprint, branch_name, commit, blueprint_name},
        _from,
        %__MODULE{dir: dir, vcli: vcli} = state
      ) do
    %Porcelain.Result{err: nil, out: _, status: 0} =
      Porcelain.exec("git", ["checkout", commit], dir: dir)

    %Porcelain.Result{err: nil, out: _, status: 0} =
      Porcelain.exec("git", ["clean", "-fd"], dir: dir)

    blueprint_plan = VCLI.plan_blueprint(dir, vcli, branch_name, blueprint_name)

    {:reply, blueprint_plan, state}
  end

  @impl true
  def handle_info(
        :fetch,
        %__MODULE__{address: address, private_key: private_key, dir: dir} = state
      ) do
    Git.Repository.Remote.fetch(address, private_key, dir)
    {:noreply, %{state | fetched: true}}
  end
end
