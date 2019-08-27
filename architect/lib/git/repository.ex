defmodule Git.Repository do
  @moduledoc """
  A process for interacting with a git repository

  On init this will immediately succeed, but trigger the `handle_continue` message as the first message handler.
  handle_continue will verify the repository in a new temporary directory.

  When this process is killed this directory should be automatically removed
  """

  defstruct [:local_path, :address, :private_key, :known_hosts, :verified, :fetched]

  defmacro __using__(_opts) do
    IO.puts("You are USING Git.Repository")
  end

  use GenServer
  require Logger
  alias Git.{Branch, Commit, Repository.Remote}

  def start_link({process_name, address, private_key, known_hosts}) do
    Logger.debug("Starting repository process for #{address}")

    GenServer.start_link(
      __MODULE__,
      {address, private_key, known_hosts},
      name: process_name,
      timeout: 10_000
    )
  end

  def initialise_with_remote(address, local_path) do
    {_out, 0} = System.cmd("git", ["init"], cd: local_path, stderr_to_stdout: true)

    {_out, 0} =
      System.cmd("git", ["remote", "add", "origin", address],
        cd: local_path,
        stderr_to_stdout: true
      )
  end

  @impl true
  def init({address, private_key, known_hosts}) do
    Temp.track!()
    local_path = Temp.mkdir!()

    {:ok,
     %__MODULE__{
       address: address,
       private_key: private_key,
       local_path: local_path,
       known_hosts: known_hosts,
       verified: false
     }, {:continue, :setup}}
  end

  @impl true
  def handle_continue(:setup, s) do
    :ok = File.mkdir_p!(s.local_path)

    with {_out, 0} <- __MODULE__.initialise_with_remote(s.address, s.local_path),
         :ok <- Remote.verify(s.address, s.private_key, s.known_hosts, s.local_path) do
      send(self(), :fetch)

      {:noreply, %__MODULE__{s | verified: true}}
    else
      {out, code} ->
        Logger.error(
          "Failed verifying git repository",
          address: s.address,
          local_path: s.local_path,
          known_hosts: s.known_hosts,
          exit_code: code,
          stdout: out
        )

        {:noreply, %__MODULE__{s | verified: false}}

      err ->
        Logger.error(err)
    end
  end

  #
  # Client
  #

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

  #
  # Server
  #

  @impl true
  def handle_info(:fetch, state) do
    :ok = Remote.fetch(state.address, state.private_key, state.known_hosts, state.local_path)
    {:noreply, %{state | fetched: true}}
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
  def handle_call(:verified, _from, state) do
    {:reply, state.verified, state}
  end

  @impl true
  def handle_call(:list_branches, _from, %__MODULE__{local_path: dir} = state) do
    branches = Git.Branch.list(dir)

    {:reply, branches, state}
  end

  def handle_call({:get_branch, branch}, _from, %__MODULE__{local_path: dir} = state) do
    branch =
      Git.Branch.list(dir)
      |> Enum.find(fn %Git.Branch{name: b} -> b == branch end)

    {:reply, branch, state}
  end

  @impl true
  def handle_call({:list_branches_for_commit, sha}, _from, %__MODULE__{local_path: dir} = state) do
    branches = Git.Branch.list_for_commit_sha(dir, sha)

    {:reply, branches, state}
  end

  @impl true
  def handle_call({:list_commits, branch}, _from, %__MODULE{local_path: dir} = state) do
    commits = Commit.list_for_ref(dir, branch)

    {:reply, commits, state}
  end

  @impl true
  def handle_call({:get_commit_by_sha, sha}, _from, %__MODULE{local_path: dir} = state) do
    commit = Commit.get_by_sha(dir, sha)

    {:reply, commit, state}
  end

  @impl true
  def handle_call(:default_branch, _from, %__MODULE__{local_path: dir} = state) do
    branch = Branch.default(dir)

    {:reply, branch, state}
  end

  @impl true
  def handle_call({:commit_count, branch}, _from, %__MODULE__{local_path: dir} = state) do
    count = Commit.count_for_branch(dir, branch)

    {:reply, count, state}
  end

  @impl true
  def handle_call(:commit_count, _from, %__MODULE__{local_path: dir} = state) do
    count = Commit.count(dir)

    {:reply, count, state}
  end
end
