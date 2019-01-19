defmodule Architect.Projects.Repository do
  @moduledoc """
  A process for interacting with a git repository

  On init this will immediately succeed, but trigger the `handle_continue` message as the first message handler.
  handle_continue will clone the repository to a new temporary directory.

  When this process is killed this directory should be automatically removed
  """

  defstruct [:status, :url, :repo]

  use GenServer
  require Logger
  alias Architect.Projects.{Branch, Commit}

  # Client

  def start_link({url, name}) when is_binary(url) do
    Logger.debug("Starting fresh repository process for #{url}")

    GenServer.start_link(__MODULE__, url, name: name)
  end

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
  Get the default branch
  """
  def default_branch(repository), do: GenServer.call(repository, :default_branch)

  # Server (callbacks)

  @impl true
  def init(url) when is_binary(url) do
    {:ok, %__MODULE__{url: url}, {:continue, :clone}}
  end

  @impl true
  def handle_continue(:clone, %__MODULE__{url: url} = state) when is_binary(url) do
    Temp.track!()

    path = Temp.mkdir!(UUID.uuid4())

    case Git.clone([url, path]) do
      {:ok, repo} ->
        Logger.debug("Successfully cloned #{url} to #{path}")
        {:noreply, %__MODULE__{state | repo: repo}}

      {:error, %Git.Error{message: reason}} ->
        Logger.error("Failed cloning #{url} to #{path}: #{reason}")
        {:stop, reason, url}
    end
  end

  @impl true
  def handle_call(:fetch, _from, %__MODULE__{repo: repo} = state) do
    Logger.debug("Performing 'fetch' on #{inspect(repo)}")

    {:ok, _} = Git.fetch(repo, ["--prune"])
    {:reply, :ok, state}
  end

  @impl true
  def handle_call(:list_branches, _from, %__MODULE__{repo: repo} = state) do
    Logger.debug("Performing 'branch' on #{inspect(repo)}")

    branches =
      repo
      |> Git.branch(["--remote"])
      |> Branch.parse()

    {:reply, branches, state}
  end

  @impl true
  def handle_call({:list_commits, branch}, _from, %__MODULE{repo: repo} = state)
      when is_binary(branch) do
    Logger.debug("Performing 'log' on #{inspect(repo)}")

    with {:ok, _} <- Git.checkout(repo, branch),
         {:ok, output} = Git.log(repo, ["--format=#{Commit.format()}"]) do
      commits = Commit.parse(output)

      {:reply, commits, state}
    end
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
end
