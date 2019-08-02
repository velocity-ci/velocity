defmodule Architect.Builds do
  @moduledoc """
  The Builds context.
  """

  require Logger
  import Ecto.Query, warn: false
  alias Architect.Repo

  alias Architect.Builds.Build
  alias Architect.Builds.Task
  alias Architect.Builds.Stage

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
  Returns the list of ready tasks.
  """
  def list_ready_tasks do
    Repo.all(
      from t in Task,
        where: t.status == "ready"
    )
  end

  @doc """
  Returns the list of building tasks.
  """
  def list_building_tasks do
    Repo.all(
      from t in Task,
        where: t.status == "building"
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
    |> Build.create_changeset(%{
      project_id: project.id,
      branch_name: branch_name,
      commit_sha: commit_sha,
      task_name: task_name,
      parameters: parameters,
      created_by_id: user.id,
      status: "building"
#      status: "waiting"
    })
    |> Repo.insert()
  end

  defp task_complete(task) do
    stage =
      Ecto.assoc(task, :stage)
      |> Repo.one()
      |> Repo.preload(:tasks)

    task_statuses =
      Enum.map(stage.tasks, fn task ->
        task.status
      end)

    stage_complete = Enum.all?(task_statuses, fn x -> x == "failed" or x == "succeeded" end)

    case stage_complete do
      true ->
        stage_complete(stage, task_statuses)

      _ ->
        :ok
    end
  end

  defp stage_complete(stage, task_statuses) do
    stage_succeeded = Enum.all?(task_statuses, fn x -> x == "succeeded" end)
    state = if stage_succeeded, do: "succeeded", else: "failed"

    Stage.update_changeset(stage, %{status: state})
    |> Repo.update()

    build =
      Ecto.assoc(stage, :build)
      |> Repo.one()
      |> Repo.preload(:stages)

    cond do
      length(build.stages) <= stage.index + 1 ->
        build_complete(build, state)

      true ->
        next_stage =
          build.stages
          |> Enum.at(stage.index + 1)

        Enum.each(next_stage.tasks, fn task ->
          Task.update_changeset(task, %{status: "ready"})
          |> Repo.update()
        end)
    end
  end

  defp build_complete(build, state) do
    Build.update_changeset(build, %{status: state})
    |> Repo.update()
  end

  def update_task(id, status) do
    {:ok, task} =
      Repo.get!(Task, id)
      |> Task.update_changeset(%{status: status})
      |> Repo.update()

    case status do
      status when status in ["failed", "succeeded"] ->
        task_complete(task)

      _ ->
        :ok
    end
  end
end
