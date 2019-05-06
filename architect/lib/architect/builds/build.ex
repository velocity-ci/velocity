defmodule Architect.Builds.Build do
  @moduledoc """
  # Notes
  ## Important differences from PoC:
  ### We don't store Git History or Tasks in the Architect's data-store:
    * A build belongs to a Task, which belongs to a Commit, but we don't want to lose the Build if say the commit gets deleted on the repository.
    * *=compound key
    Build:
      UUID:
      *Project: FK -> projects table
      *Commit: store in JSON? how do we query with ecto? or do we just create records for ones that are built. Or we store the Hash and query the source code repo (returning Not found warning if missing/deleted)
      *Task: ""
      *CreatedAt

  """
  use Ecto.Schema
  import Ecto.Changeset
  alias Ecto.Changeset
  require Logger
  alias Architect.Accounts.User
  alias Architect.Projects.Project

  @primary_key {:id, :binary_id, autogenerate: true}
  @foreign_key_type :binary_id
  schema "builds" do
    belongs_to(:project, Project)
    field(:branch_name, :string)
    field(:commit_sha, :string)
    field(:task_name, :string)
    field(:parameters, :map)

    field(:plan, :map)

    field(:status, :string)
    field(:created_at, :utc_datetime)
    field(:updated_at, :utc_datetime)
    field(:started_at, :utc_datetime)
    field(:completed_at, :utc_datetime)

    belongs_to(:created_by, User)
  end

  @doc false
  def changeset(%__MODULE__{} = build, attrs) do
    # Check repository has task_name in commit_sha and task is 'valid' (vcli)
    build
    |> cast(attrs, [
      :project_id,
      :branch_name,
      :commit_sha,
      :task_name,
      :parameters,
      :created_by_id,
      :status,
    ])
    |> assoc_constraint(:project)
    |> assoc_constraint(:created_by)
    |> validate_required([:task_name, :commit_sha])
    |> set_plan()
  end

  def set_plan(
    %Changeset{
      valid?: true,
      changes: %{
        task_name: task_name,
        commit_sha: commit_sha,
        project_id: project_id,
      }
    } = changeset
  ) do
    project = Architect.Projects.get_project!(project_id)
    plan = Architect.Projects.plan_blueprint(project, commit_sha, task_name)
    changeset
    |> put_change(:plan, plan)
    |> put_change(:id, plan["id"])
  end

  def set_plan(changeset), do: changeset
end
