defmodule Architect.Builds do
  @moduledoc """
  The Builds context.
  """

  import Ecto.Query, warn: false
  alias Architect.Repo

  alias Architect.Builds.Build

  @doc """
  Returns the list of builds.
  """
  def list_builds do
    Repo.all(
      from b in Build,
        join: p in assoc(b, :project),
        preload: [project: p]
    )
  end

  @doc """
  Returns the list of waiting builds.
  """
  def list_waiting_builds do
    Repo.all(
      from b in Build,
        where: b.status == "waiting",
        join: p in assoc(b, :project),
        preload: [project: p]
    )
  end

  @doc """
  Returns the list of running builds.
  """
  def list_running_builds do
    Repo.all(
      from b in Build,
        where: b.status == "running",
        join: p in assoc(b, :project),
        preload: [project: p]
    )
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
      status: "waiting"
    })
    |> Repo.insert()

    # TODO: put tasks from construction plan into ETS
  end
end
