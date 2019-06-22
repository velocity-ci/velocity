defmodule Architect.Builds.Task do
  @moduledoc """

  """

  use Ecto.Schema
  import Ecto.Changeset
  alias Architect.Builds.Build

  @primary_key {:id, :binary_id, autogenerate: true}
  @foreign_key_type :binary_id
  schema "tasks" do
    belongs_to(:build, Build)

    field(:plan, :map)

    field(:status, :string, default: "waiting")
    field(:created_at, :utc_datetime)
    field(:updated_at, :utc_datetime)
    field(:started_at, :utc_datetime)
    field(:completed_at, :utc_datetime)
  end

  @doc false
  def changeset(%__MODULE__{} = task, attrs) do
    task
    |> cast(attrs, [
      :plan
    ])
    |> assoc_constraint(:build)
    |> validate_required([:plan])
  end
end
