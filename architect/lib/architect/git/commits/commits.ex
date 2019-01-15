defmodule Architect.Git.Commits do
  @moduledoc """
  The Commits context.
  """

  import Ecto.Query, warn: false
  alias Architect.Repo

  alias Architect.Git.Commits.Commit

  @doc """
  Returns the list of commits.

  ## Examples

      iex> list_commits()
      [%Commit{}, ...]

  """
  def list_commits do
    Repo.all(Commit)
  end

  @doc """
  Gets a single commit by id.

  Raises `Ecto.NoResultsError` if the Known host does not exist.

  ## Examples

      iex> get_commit!(123)
      %Commit{}

      iex> get_commit!(456)
      ** (Ecto.NoResultsError)

  """
  def get_commit!(id), do: Repo.get!(Commit, id)

  @doc """
  Creates a commit.

  ## Examples

      iex> create_commit(%{field: value})
      {:ok, %Commit{}}

      iex> create_commit(%{field: bad_value})
      {:error, %Ecto.Changeset{}}

  """
  def create_commit(project, attrs \\ %{}) do
    %Commit{
      project_id: project.id
    }
    |> Commit.changeset(attrs)
    |> Repo.insert()
  end

  @doc """
  Updates a commit.

  ## Examples

      iex> update_commit(commit, %{field: new_value})
      {:ok, %Commit{}}

      iex> update_commit(commit, %{field: bad_value})
      {:error, %Ecto.Changeset{}}

  """
  def update_commit(%Commit{} = commit, attrs) do
    commit
    |> Commit.changeset(attrs)
    |> Repo.update()
  end

  @doc """
  Deletes a Commit.

  ## Examples

      iex> delete_commit(commit)
      {:ok, %Commit{}}

      iex> delete_commit(commit)
      {:error, %Ecto.Changeset{}}

  """
  def delete_commit(%Commit{} = commit) do
    Repo.delete(commit)
  end

  @doc """
  Returns an `%Ecto.Changeset{}` for tracking commit changes.

  ## Examples

      iex> change_commit(commit)
      %Ecto.Changeset{source: %Commit{}}

  """
  def change_commit(%Commit{} = commit) do
    Commit.changeset(commit, %{})
  end
end
