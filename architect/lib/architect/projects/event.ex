defmodule Architect.Projects.Event do
  import EctoEnum
  use Ecto.Schema
  import Ecto.Changeset

  defenum(TypeEnum, :project_event_type, [:created, :updated, :deleted])

  alias Architect.Projects.Project
  alias Architect.Accounts.User

  @primary_key {:id, :binary_id, autogenerate: true}
  @foreign_key_type :binary_id
  schema "project_events" do
    field(:type, TypeEnum)
    field(:metadata, :map)

    belongs_to(:project, Project)
    belongs_to(:user, User)

    timestamps()
  end

  @doc false
  def changeset(%__MODULE__{} = event, attrs) do
    event
    |> cast(attrs, [:type, :metadata, :project_id, :user_id])
    |> validate_required([:type])
    |> assoc_constraint(:project)
    |> assoc_constraint(:user)
  end
end
