defmodule Architect.Projects.Repository do
  @moduledoc """
  A process for interacting with a git repository

  On init this will immediately succeed, but trigger the `handle_continue` message as the first message handler.
  handle_continue will clone the repository to a new temporary directory.

  When this process is killed this directory should be automatically removed
  """

  defstruct [:status, :url, :repo, :vcli, :commits_table, :branches_table, :commit_branches_table]

  use GenServer
  require Logger
  alias Architect.Projects.{Branch, Commit, Task}
  alias Architect.VCLI

  @timeout 20_000

  # Client

  def start_link({url, name}) when is_binary(url) do
    Logger.debug("Starting fresh repository process for #{url}")

    GenServer.start_link(__MODULE__, url, name: name, timeout: 10_000)
  end

  @doc ~S"""
  Check if cloned successfully
  """
  def cloned_successfully?(repository),
    do: GenServer.call(repository, :clone_status, @timeout) == :cloned

  @doc ~S"""
  Get commit amount across all branches
  """
  def commit_count(repository), do: GenServer.call(repository, :commit_count, @timeout)

  @doc ~S"""
  Get commit amount for branch
  """
  def commit_count(repository, branch),
    do: GenServer.call(repository, {:commit_count, branch}, @timeout)

  @doc ~S"""
  Run `git fetch` on the repository
  """
  def fetch(repository), do: GenServer.call(repository, :fetch, @timeout)

  @doc ~S"""
  Get a single commit by its SHA value
  """
  def commit_by_sha(repository, sha),
    do: GenServer.call(repository, {:get_commit_by_sha, sha}, @timeout)

  @doc ~S"""
  Get a list of commits by branch
  """
  def list_commits(repository, branch),
    do: GenServer.call(repository, {:list_commits, branch}, @timeout)

  @doc ~S"""
  Get a list of branches
  """
  def list_branches(repository), do: GenServer.call(repository, :list_branches, @timeout)

  @doc ~S"""
  Get a list of branches for a commit SHA value
  """
  def list_branches_for_commit(repository, sha),
    do: GenServer.call(repository, {:list_branches_for_commit, sha}, @timeout)

  @doc ~S"""
  Get the default branch
  """
  def default_branch(repository), do: GenServer.call(repository, :default_branch, @timeout)

  @doc ~S"""
  Ge the tasks for a commit specified by its SHA value
  """
  def list_tasks(repository, selector),
    do: GenServer.call(repository, {:list_tasks, selector}, @timeout)

  # Server (callbacks)

  @impl true
  def init(url) when is_binary(url) do
    {:ok, %__MODULE__{url: url, status: :cloning, vcli: VCLI.init()}, {:continue, :clone}}
  end

  @impl true
  def handle_continue(:clone, %__MODULE__{url: url} = state) when is_binary(url) do
    Logger.debug("Creating ETS tables for #{url}")

    #    branches_table = :ets.new(:branches, [:set, :private])
    #    commits_table = :ets.new(:commits, [:set, :private])
    #    commit_branches_table = :ets.new(:commit_branches, [:set, :private])

    Temp.track!()

    path = Temp.mkdir!(UUID.uuid4())

    case Git.clone([url, path]) do
      {:ok, repo} ->
        Logger.debug("Successfully cloned #{url} to #{path}")

        branches =
          repo
          |> Git.branch()
          |> Branch.parse()

        for branch <- branches do
          {:ok, _output} = Git.checkout(repo, [branch.name])
        end

        {:noreply,
         %__MODULE__{
           state
           | repo: repo,
             status: :cloned
             #             commits_table: commits_table,
             #             branches_table: branches_table,
             #             commit_branches_table: commit_branches_table
         }}

      {:error, %Git.Error{message: reason}} ->
        Logger.error("Failed cloning #{url} to #{path}: #{reason}")
        {:noreply, %__MODULE__{state | status: :failed}}
    end
  end

  @impl true
  def handle_call(:clone_status, _from, %__MODULE__{status: status} = state) do
    {:reply, status, state}
  end

  @impl true
  def handle_call(_, _from, %__MODULE__{status: :failed} = state) do
    Logger.warn("Cannot perform action on failed repository")

    {:reply, :error, state}
  end

  @impl true
  def handle_call(:fetch, _from, %__MODULE__{repo: repo} = state) do
    Logger.debug("Performing 'fetch' on #{inspect(repo)}")

    {:ok, _} = Git.fetch(repo, ["--prune"])
    {:reply, :ok, state}
  end

  @impl true
  def handle_call(
        :list_branches,
        _from,
        %__MODULE__{repo: repo, branches_table: branches_table} = state
      ) do
    Logger.debug("Performing 'branch' on #{inspect(repo)}")

    branches =
      repo
      |> Git.branch()
      |> Branch.parse()

    #    :ets.insert(branches_table, branches)

    {:reply, branches, state}
  end

  @impl true
  def handle_call(
        {:list_branches_for_commit, sha},
        _from,
        %__MODULE__{repo: repo, commit_branches_table: commit_branches_table} = state
      ) do
    Logger.debug("Performing 'branch --contains' on #{inspect(repo)}")

    branches =
      repo
      |> Git.branch(["--contains=#{sha}"])
      |> Branch.parse()

    #    :ets.insert(commit_branches_table, {sha, branches})

    {:reply, branches, state}
  end

  @impl true
  def handle_call({:list_commits, branch}, _from, %__MODULE{repo: repo} = state)
      when is_binary(branch) do
    Logger.debug("Performing 'checkout #{branch}}' on #{inspect(repo)}")

    {:ok, _} = Git.checkout(repo, branch)

    Logger.debug("Performing 'log --format=#{Commit.format()}}' on #{inspect(repo)}")

    {:ok, output} = Git.log(repo, ["--format=#{Commit.format()}", branch])

    Logger.debug("Parsing commit on #{inspect(repo)}")

    commits = Commit.parse(output)

    {:reply, commits, state}
  end

  @impl true
  def handle_call({:get_commit_by_sha, sha}, _from, %__MODULE{repo: repo} = state)
      when is_binary(sha) do
    Logger.debug("Performing 'show' on #{inspect(repo)}")

    commit =
      repo
      |> Git.show(["-s", "--format=#{Commit.format()}", sha])
      |> Commit.parse_show()

    {:reply, commit, state}
  end

  @impl true
  def handle_call(:default_branch, _from, %__MODULE__{repo: repo} = state) do
    Logger.debug("Performing 'remote show origin' on #{inspect(repo)}")

    branch =
      repo
      |> Git.remote(["show", "origin"])
      |> Branch.parse_remote()

    {:reply, branch, state}
  end

  @impl true
  def handle_call({:commit_count, branch}, _from, %__MODULE__{repo: repo} = state)
      when is_binary(branch) do
    Logger.debug(
      "Performing 'remote show origin' on #{inspect(repo)} for branch #{inspect(branch)}"
    )

    {:ok, _} = Git.checkout(repo, [branch, "--force"])

    count =
      repo
      |> Git.rev_list(["--count", branch])
      |> Commit.parse_count()

    {:reply, count, state}
  end

  @impl true
  def handle_call(:commit_count, _from, %__MODULE__{repo: repo} = state) do
    Logger.debug("Performing 'remote show origin' on #{inspect(repo)}")

    count =
      repo
      |> Git.rev_list(["--count", "--all"])
      |> Commit.parse_count()

    {:reply, count, state}
  end

  @impl true
  def handle_call(
        {:list_tasks, {:sha, sha}},
        _from,
        %__MODULE__{repo: repo, vcli: vcli} = state
      ) do
    Logger.debug("Performing VCLI cmd on #{inspect(repo)} for SHA #{inspect(sha)}}")

    {:ok, _output} = Git.checkout(repo, [sha, "--force"])

    tasks =
      vcli
      |> VCLI.list()
      |> Enum.map(&Task.parse/1)

    {:reply, tasks, state}
  end

  @impl true
  def handle_call(
        {:list_tasks, {:branch, branch}},
        _from,
        %__MODULE__{repo: repo, vcli: vcli} = state
      ) do
    Logger.debug("Performing VCLI cmd on #{inspect(repo)} for branch #{inspect(branch)}}")

    {:ok, _output} = Git.checkout(repo, [branch, "--force"])

    tasks =
      vcli
      |> VCLI.list()
      |> Enum.map(&Task.parse/1)

    {:reply, tasks, state}
  end
end
