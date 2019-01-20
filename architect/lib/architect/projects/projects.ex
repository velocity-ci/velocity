defmodule Architect.Projects do
  @moduledoc """
  The Projects context.
  """

  import Ecto.Query, warn: false
  alias Architect.Repo
  alias Architect.Projects.{Project, Repository, Starter}
  use Supervisor
  require Logger

  @registry __MODULE__.Registry
  @supervisor __MODULE__.Supervisor

  def start_link(_opts \\ []) do
    Logger.debug("Starting #{Atom.to_string(__MODULE__)}")
    Supervisor.start_link(__MODULE__, :ok, name: __MODULE__)
  end

  @doc """
  Returns the list of projects.

  ## Examples

      iex> list_projects()
      [%Project{}, ...]

  """
  def list_projects() do
    Repo.all(Project)
  end

  @doc """
  Gets a single project by id.

  Raises `Ecto.NoResultsError` if the Known host does not exist.
  """
  def get_project!(id), do: Repo.get!(Project, id)

  @doc """
  Gets a single project by slug.

  Raises `Ecto.NoResultsError` if the Project does not exist.

  ## Examples

      iex> get_project_by_slug!("velocity")
      %KnownHost{}

      iex> get_project_by_slug!("Not a slug")
      ** (Ecto.NoResultsError)

  """
  def get_project_by_slug!(slug), do: Repo.get_by!(Project, slug: slug)

  @doc """
  Creates a project.

  ## Examples

      iex> create_project(%{field: value})
      {:ok, %Project{}}

      iex> create_project(%{field: bad_value})
      {:error, %Ecto.Changeset{}}

  """
  def create_project(attrs \\ %{}) do
    %Project{}
    |> Project.changeset(attrs)
    |> Repo.insert()
  end

  @doc ~S"""
  Get a list of branches

  ## Examples

      iex> list_branches(project)
      [%Branch{}, ...]

  """
  def list_branches(%Project{} = project),
    do: call_repository(project, &Repository.list_branches/1)

  @doc ~S"""
  Get a list of branches for a specific commit SHA

  ## Examples

      iex> list_branches_for_commit("925fbc450c8bdb7665ec3af3129ce715927433fe")
      [%Branch{}, ...]

  """
  def list_branches_for_commit(%Project{} = project, sha) when is_binary(sha),
    do: call_repository(project, &Repository.list_branches_for_commit(&1, sha))

  @doc ~S"""
  Get a list of commits by branch


  ## Examples

      iex> list_commits(project, "master")
      [%Commit{}, ...]

  """
  def list_commits(%Project{} = project, branch) when is_binary(branch),
    do: call_repository(project, &Repository.list_commits(&1, branch))

  @doc ~S"""
  Get the default branch

  ## Examples

      iex> default_branch(project)
      %Branch{}

  """
  def default_branch(%Project{} = project),
    do: call_repository(project, &Repository.default_branch/1)

  @doc ~S"""
  Get the amount of commits for the project

  ## Examples

      iex> commit_count(project)
      123

  """
  def commit_count(%Project{} = project),
    do: call_repository(project, &Repository.commit_count/1)

  @doc ~S"""
  Get the amount of commits for the project, for a specific branch

  ## Examples

      iex> commit_count(project, "master")
      42

  """
  def commit_count(%Project{} = project, branch) when is_binary(branch),
    do: call_repository(project, &Repository.commit_count(&1, branch))

  ### Server

  @impl true
  def init(:ok) do
    children = [
      {Registry, keys: :unique, name: @registry},
      {DynamicSupervisor, name: @supervisor, strategy: :one_for_one, max_restarts: 3},
      worker(
        Starter,
        [%{registry: @registry, supervisor: @supervisor, projects: list_projects()}],
        restart: :transient
      )
    ]

    Logger.info("Running #{Atom.to_string(__MODULE__)}")

    Supervisor.init(children, strategy: :one_for_one)
  end

  @doc false
  defp call_repository(project, callback, attempt \\ 1)

  defp call_repository(%Project{address: address} = project, callback, attempt)
       when attempt < 3 do
    case Registry.lookup(@registry, address) do
      [{repository, _}] ->
        try do
          callback.(repository)
        catch
          kind, reason ->
            Logger.warn(
              "Failed to call repository #{address} (#{inspect(kind)} #{inspect(reason)}), retrying..."
            )

            Process.sleep(1_000)

            call_repository(project, callback, attempt + 1)
        end

      [] ->
        Logger.warn(
          "Failed to call builder #{address} on #{inspect(@registry)}; address does not exist"
        )

        {:error, "Not found"}
    end
  end

  defp call_repository(_, _, _) do
    {:error, "Failed"}
  end
end

defmodule Architect.Projects.Starter do
  use Task
  require Logger
  alias Architect.Projects.Repository

  def start_link(opts) do
    Task.start_link(__MODULE__, :run, [opts])
  end

  def run(%{projects: projects, supervisor: supervisor, registry: registry}) do
    for project <- projects do
      {:ok, _} =
        DynamicSupervisor.start_child(
          supervisor,
          {Repository, {project.address, {:via, Registry, {registry, project.address}}}}
        )
    end
  end

  def run([]), do: nil
end
