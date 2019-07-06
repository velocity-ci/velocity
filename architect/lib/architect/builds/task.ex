defmodule Architect.Builds.Task do
  @moduledoc """

  """

  use Ecto.Schema
  import Ecto.Changeset
  alias Architect.Builds.Stage

  @primary_key {:id, :binary_id, autogenerate: true}
  @foreign_key_type :binary_id
  schema "tasks" do
    belongs_to(:stage, Stage)

    field(:plan, :map)

    field(:status, :string, default: "waiting")
    field(:created_at, :utc_datetime)
    field(:updated_at, :utc_datetime)
    field(:started_at, :utc_datetime)
    field(:completed_at, :utc_datetime)
  end

  @doc false
  def create_changeset(%__MODULE__{} = task, attrs) do
    task
    |> cast(attrs, [
      :plan,
      :status
    ])
    |> assoc_constraint(:stage)
    |> validate_required([:plan])
  end

  @doc false
  def update_changeset(%__MODULE__{} = build, attrs) do
    build
    |> cast(attrs, [
      :plan,
      :status
    ])
  end
end
