defmodule Architect.Events.Event do
  import EctoEnum
  use Ecto.Schema
  import Ecto.Changeset

  defenum(TypeEnum, :event_type, [:project_created, :known_host_created, :known_host_verified])

  alias Architect.Projects.Project
  alias Architect.Accounts.User
  alias Architect.KnownHosts.KnownHost

  @primary_key {:id, :binary_id, autogenerate: true}
  @foreign_key_type :binary_id
  schema "events" do
    field(:type, TypeEnum)
    field(:metadata, :map)

    belongs_to(:project, Project)
    belongs_to(:known_host, KnownHost)
    belongs_to(:user, User)

    timestamps()
  end

  @doc false
  def changeset(%__MODULE__{} = event, attrs) do
    event
    |> cast(attrs, [:type, :metadata, :project_id, :known_host_id, :user_id])
    |> validate_required([:type])
    |> assoc_constraint(:user)
  end
end
