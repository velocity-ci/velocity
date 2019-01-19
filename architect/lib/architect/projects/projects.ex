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
  Get a list of commits by branch


  ## Examples

      iex> list_commits(project, "master")
      [%Commit{}, ...]

  """
  def list_commits(%Project{} = project, branch),
    do: call_repository(project, &Repository.list_commits(&1, branch))

  @doc ~S"""
  Get the default branch

  ## Examples

      iex> default_branch(project)
      %Branch{}

  """
  def default_branch(%Project{} = project),
    do: call_repository(project, &Repository.default_branch/1)

  ### Server

  @impl true
  def init(:ok) do
    children = [
      {Registry, keys: :unique, name: @registry},
      {DynamicSupervisor, name: @supervisor, strategy: :one_for_one},
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
  defp call_repository(%Project{id: id}, callback) do
    case Registry.lookup(@registry, id) do
      [{repository, _}] ->
        try do
          callback.(repository)
        catch
          kind, reason ->
            formatted = Exception.format(kind, reason, __STACKTRACE__)

            Logger.error(
              "Failed to call repository #{id} on #{inspect(@registry)} with #{formatted}"
            )
        end

      [] ->
        Logger.error("Failed to call builder #{id} on #{inspect(@registry)}; id does not exist")

        {:error, :not_found}
    end
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
          {Repository, {project.address, {:via, Registry, {registry, project.id}}}}
        )
    end
  end

  def run([]), do: nil
end
