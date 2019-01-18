defmodule Architect.Projects.Repository do
  @moduledoc """
  A process for interacting with a git repository
  """

  defstruct [:status, :url, :repo]

  use GenServer
  require Logger
  alias Architect.Projects.{Branch, Commit}

  # Client

  def start_link({url, _name}) when is_binary(url) do
    Logger.debug("Starting fresh repository process for #{url}")

    GenServer.start_link(__MODULE__, url)
  end

  def fetch(repository), do: GenServer.call(repository, :fetch)

  def commit_by_sha(repository, sha), do: GenServer.call(repository, {:get_commit_by_sha, sha})

  def list_commits(repository, branch), do: GenServer.call(repository, {:list_commits, branch})

  def list_branches(repository), do: GenServer.call(repository, :list_branches)

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

    {:ok, output} = Git.branch(repo, ["--remote"])
    branches = Branch.parse(output)

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

    {:ok, output} = Git.show(repo, ["-s", "--format=#{Commit.format()}", sha])

    commit = Commit.parse_show(output)

    {:reply, commit, state}
  end

  @impl true
  def handle_call(:default_branch, _from, %__MODULE__{repo: repo} = state) do
    Logger.debug("Performing 'remote show origin' on #{inspect(repo)}")

    {:ok, output} = Git.remote(repo, ["show", "origin"])
    branch = Branch.parse_remote(output)

    {:reply, branch, state}
  end
end
