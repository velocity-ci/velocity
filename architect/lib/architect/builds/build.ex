defmodule Architect.Builds.Build do
  @moduledoc """

  """

  use Ecto.Schema
  import Ecto.Changeset
  alias Ecto.Changeset
  require Logger
  alias Architect.Accounts.User
  alias Architect.Projects.Project
  alias Architect.Builds.Stage

  @primary_key {:id, :binary_id, autogenerate: true}
  @foreign_key_type :binary_id
  schema "builds" do
    belongs_to(:project, Project)
    field(:branch_name, :string)
    field(:commit_sha, :string)
    field(:task_name, :string)
    field(:parameters, :map)

    field(:status, :string, default: "building")
    field(:created_at, :utc_datetime)
    field(:updated_at, :utc_datetime)
    field(:started_at, :utc_datetime)
    field(:completed_at, :utc_datetime)

    has_many(:stages, Stage)
    belongs_to(:created_by, User)
  end

  @doc false
  def create_changeset(%__MODULE__{} = build, attrs) do
    # Check repository has task_name in commit_sha and task is 'valid' (vcli)
    build
    |> cast(attrs, [
      :project_id,
      :branch_name,
      :commit_sha,
      :task_name,
      :parameters,
      :created_by_id,
      :status
    ])
    |> assoc_constraint(:project)
    |> assoc_constraint(:created_by)
    |> validate_required([:task_name, :commit_sha])
    |> set_plan()
  end

  @doc false
  def update_changeset(%__MODULE__{} = build, attrs) do
    build
    |> cast(attrs, [
      :status
    ])
  end

  def set_plan(
        %Changeset{
          valid?: true,
          changes: %{
            task_name: task_name,
            commit_sha: commit_sha,
            project_id: project_id,
            branch_name: branch_name
          }
        } = changeset
      ) do
    project = Architect.Projects.get_project!(project_id)
    plan = Architect.Projects.plan_blueprint(project, branch_name, commit_sha, task_name)

    stages =
      Enum.map(plan["stages"], fn stage ->
        %Stage{}
        |> Stage.create_changeset(stage)
      end)

    changeset
    |> put_change(:id, plan["id"])
    |> put_assoc(:stages, stages)
  end

  def set_plan(changeset), do: changeset
end

defmodule Architect.Builds.Step do
  use Ecto.Schema

  embedded_schema do
    field(:status, :string)
    field(:updated_at, :utc_datetime)
    field(:started_at, :utc_datetime)
    field(:completed_at, :utc_datetime)

    embeds_one(:streams, {:map, Architect.Builds.Stream})
  end
end

defmodule Architect.Builds.Stream do
  use Ecto.Schema

  embedded_schema do
    field(:status, :string)
    field(:updated_at, :utc_datetime)
    field(:started_at, :utc_datetime)
    field(:completed_at, :utc_datetime)

    # Store the sources of streams rather than contents so we don't store actual logs in the database
    # Can do plugins for schemes like ets://, file://, s3://, dynamodb://, mongodb:// etc.
    field(:source, :string)
  end
end
