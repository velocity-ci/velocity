defmodule Velocity.UserValidator do
  @moduledoc "Provides user validation"

  use Ecto.Schema
  import Ecto.Changeset

  @spec register(Ecto.Changeset) :: {}
  def register(changeset) do
    changeset = changeset
      |> validate_required([:username, :password])
      |> validate_length(:password, min: 8)
      |> validate_length(:username, min: 3)
      |> validate_format(:username, ~r/^\S*$/, message: "cannot contain spaces")
    if changeset.valid? do
      {:ok, changeset}
    else
      {:error, changeset}
    end
  end

  @spec login(Ecto.Changeset) :: {}
  def login(changeset) do
    changeset = changeset
      |> validate_required([:username, :password])
    if changeset.valid? do
      {:ok, changeset}
    else
      {:error, changeset}
    end
  end
end
