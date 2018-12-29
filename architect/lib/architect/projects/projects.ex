defmodule Architect.Projects do
  @moduledoc """
  The Projects context.
  """

  import Ecto.Query, warn: false
  alias Architect.Repo

  alias Architect.Projects.Project

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
end
