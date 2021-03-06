defmodule Architect.Builds.Stage do
  @moduledoc """

  """

  use Ecto.Schema
  import Ecto.Changeset
  alias Ecto.Changeset
  alias Architect.Builds.Build
  alias Architect.Builds.Task

  @primary_key {:id, :binary_id, autogenerate: true}
  @foreign_key_type :binary_id
  schema "stages" do
    belongs_to(:build, Build)
    has_many(:tasks, Task)

    field(:index, :integer)
    field(:status, :string, default: "waiting")
    field(:created_at, :utc_datetime)
    field(:updated_at, :utc_datetime)
    field(:started_at, :utc_datetime)
    field(:completed_at, :utc_datetime)
  end

  @doc """
  """
  def create_changeset(%__MODULE__{} = stage, attrs) do
    stage
    |> cast(attrs, [
      :id,
      :index
    ])
    |> assoc_constraint(:build)
    |> validate_required([:id, :index])
    |> parse_tasks(attrs["tasks"], attrs["index"])
  end

  @doc """
  """
  def update_changeset(%__MODULE__{} = stage, attrs) do
    stage
    |> cast(attrs, [
      :status
    ])
  end

  defp parse_tasks(
         %Changeset{
           valid?: true
         } = changeset,
         tasks_json,
         i
       ) do
    tasks =
      Enum.map(tasks_json, fn {id, task} ->
        if i == 1 do
          %Task{id: id}
          |> Task.create_changeset(%{plan: task, status: "ready"})
        else
          %Task{id: id}
          |> Task.create_changeset(%{plan: task})
        end
      end)

    changeset
    |> put_assoc(:tasks, tasks)
  end
end
