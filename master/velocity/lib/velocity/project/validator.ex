defmodule Velocity.ProjectValidator do
  @moduledoc "Provides project validation"

  use Ecto.Schema
  import Ecto.Changeset

  @spec create(Ecto.Changeset) :: {}
  def create(changeset) do
    changeset = changeset
      |> validate_required([:name, :repository, :key])
      |> validate_length(:name, min: 3)
      |> validate_length(:repository, min: 8)
      |> validate_length(:key, min: 8)
    if changeset.valid? do
      {:ok, changeset}
    else
      {:error, changeset}
    end
  end
end
