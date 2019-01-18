defmodule Architect.Projects.Repository do
  @moduledoc """
  A process for interacting with a git repository
  """

  defstruct [:repo]
  use GenServer
  require Logger
  alias Architect.Projects.Branch

  # Client

  def start_link({url, _name}) when is_binary(url) do
    Logger.debug("Starting fresh repository process for #{url}")

    GenServer.start_link(__MODULE__, url)
  end

  def fetch(repository) do
    GenServer.call(repository, :fetch)
  end

  def list_branches(repository) do
    GenServer.call(repository, :list_branches)
  end

  def default_branch(repository) do
    GenServer.call(repository, :default_branch)
  end

  # Server (callbacks)

  @impl true
  def init(url) when is_binary(url) do
    Temp.track!()

    path = Temp.mkdir!(UUID.uuid4())

    with {:ok, repo} <- Git.clone([url, path]) do
      Logger.debug("Successfully cloned #{url} to #{path}")
      {:ok, %__MODULE__{repo: repo}}
    else
      {:error, %Git.Error{message: reason}} ->
        Logger.error("Failed cloning #{url} to #{path}: #{reason}")
        {:stop, reason}
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
  def handle_call(:default_branch, _from, %__MODULE__{repo: repo} = state) do
    Logger.debug("Performing 'remote show origin' on #{inspect(repo)}")

    {:ok, output} = Git.remote(repo, ["show", "origin"])
    branch = Branch.parse_remote(output)

    {:reply, branch, state}
  end
end
