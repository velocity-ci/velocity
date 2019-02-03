defmodule Architect.KnownHosts.Event do
  import EctoEnum
  use Ecto.Schema
  import Ecto.Changeset

  defenum(TypeEnum, :known_host_event_type, [:created, :updated, :deleted])

  alias Architect.KnownHosts.KnownHost
  alias Architect.Accounts.User

  @primary_key {:id, :binary_id, autogenerate: true}
  @foreign_key_type :binary_id

  schema "known_host_events" do
    field(:type, TypeEnum)
    field(:metadata, :map)

    belongs_to(:known_host, KnownHost)
    belongs_to(:user, User)

    timestamps()
  end

  @doc false
  def changeset(%__MODULE__{} = event, attrs) do
    event
    |> cast(attrs, [:type, :metadata, :known_host_id, :user_id])
    |> validate_required([:type])
    |> assoc_constraint(:known_host)
    |> assoc_constraint(:user)
  end
end
