defmodule Architect.Users.User do
  use Ecto.Schema
  import Ecto.Changeset

  alias Comeonin.Bcrypt

  @primary_key {:id, :binary_id, autogenerate: true}
  schema "users" do
    field(:username, :string)
    field(:password, :string)

    timestamps()
  end

  @doc false
  def changeset(user, attrs) do
    user
    |> cast(attrs, [:username, :password])
    |> validate_required([:username, :password])
    |> unique_constraint(:username)
    |> put_password_hash()
  end

  defp put_password_hash(
         %Ecto.Changeset{valid?: true, changes: %{password: password}} = changeset
       ) do
    change(changeset, password: Bcrypt.hashpwsalt(password))
  end

  defp put_password_hash(changeset), do: changeset
end
