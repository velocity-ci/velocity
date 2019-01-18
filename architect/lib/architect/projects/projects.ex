defmodule Architect.Projects do
  @moduledoc """
  The Projects context.
  """

  import Ecto.Query, warn: false
  alias Architect.Repo
  alias Architect.Projects.{Project, Repository, Starter}
  use Supervisor
  require Logger

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

  @doc """
  Deletes a Project.

  ## Examples

      iex> delete_project(project)
      {:ok, %Project{}}

      iex> delete_project(project)
      {:error, %Ecto.Changeset{}}

  """
  def delete_project(%Project{} = project) do
    Repo.delete(project)
  end

  ### Server

  @impl true
  def init(:ok) do
    children = [
      {Registry, keys: :unique, name: __MODULE__.RepositoryRegistry},
      {DynamicSupervisor, name: __MODULE__.RepositorySupervisor, strategy: :one_for_one},
      worker(Starter, [list_projects()], restart: :transient)
    ]

    Logger.info("Running #{Atom.to_string(__MODULE__)}")

    Supervisor.init(children, strategy: :one_for_one)
  end
end

defmodule Architect.Projects.Starter do
  use Task
  alias Architect.Projects.{Project, Repository, RepositoryRegistry, RepositorySupervisor}

  def start_link(projects) do
    Task.start_link(__MODULE__, :run, [projects])
  end

  def start_link([]), do: :ok

  def run(projects) do
    for project <- projects do
      DynamicSupervisor.start_child(
        RepositorySupervisor,
        {Repository, {project.address, {:via, Registry, {RepositoryRegistry, project.address}}}}
      )
    end
  end

  def run([]), do: nil
end
