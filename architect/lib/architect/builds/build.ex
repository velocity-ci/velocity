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
    field(:task_name, :string)
    field(:commit_sha, :string)
    field(:parameters, :map)
    field(:status, :string)

    field(:created_at, :utc_datetime)
    field(:updated_at, :utc_datetime)
    field(:started_at, :utc_datetime)
    field(:completed_at, :utc_datetime)

    belongs_to(:created_by, User)
  end

  @doc false
  def changeset(%__MODULE__{} = known_host, attrs) do
    # Check repository has task_name in commit_sha
    known_host
    |> cast(attrs, [:project_id, :task_name, :commit_sha, :parameters, :created_by_id])
    |> assoc_constraint(:created_by)
    |> validate_required([:host])
    |> unique_constraint(:host)
  end
end
