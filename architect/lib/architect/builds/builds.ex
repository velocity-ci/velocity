defmodule Architect.Builds do
  @moduledoc """
  The Builds context.
  """

  import Ecto.Query, warn: false
  alias Architect.Repo

  alias Architect.Builds.Build
  alias Comeonin.Bcrypt

  @doc """
  Returns the list of builds.
  """
  def list_builds do
    Repo.all(Build)
  end

  @doc """
  Gets a single build.

  Raises `Ecto.NoResultsError` if the Build does not exist.
  """
  def get_build!(id), do: Repo.get!(Build, id)

  @doc """
  Creates a build.
  """
  def create_build(user, project, branch_name, commit_sha, task_name, parameters \\ %{}) do
    %Build{}
    |> Build.changeset(%{
      project_id: project.id,
      branch_name: branch_name,
      commit_sha: commit_sha,
      task_name: task_name,
      parameters: parameters,
      created_by_id: user.id,
    })
    |> Repo.insert()
  end
end
