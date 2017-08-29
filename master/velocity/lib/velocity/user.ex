defmodule Velocity.User do
  @moduledoc "provides user entity"
  use Ecto.Schema
  import Ecto.Changeset
  alias Ecto.Changeset
  alias Comeonin.Bcrypt
  alias Velocity.User
  alias Velocity.Repo
  alias Velocity.UserRepository

  schema "users" do
    field :username, :string
    field :password, :string, virtual: true
    field :hashed_password, :string
    field :email, :string
    timestamps()
  end

  @doc false
  def changeset(%User{} = user, attrs) do
    user
    |> cast(attrs, [:username, :email, :password])
  end

  @spec register(Ecto.Changeset) :: struct
  def register(changeset) do
    changeset = Changeset.unique_constraint(changeset, :username)
    changeset = Changeset.unique_constraint(changeset, :email)
    case Repo.insert(changeset) do
      {:ok, user} ->
        inserted_changeset = Changeset.change(user)
        inserted_changeset
        |> put_change(:hashed_password,
                      Bcrypt.hashpwsalt(changeset.params["password"]))
        |> Repo.update

      {:error, changeset} ->
        {:error, changeset}
    end
  end

  @spec find_and_check_password(Ecto.Changeset) :: User
  def find_and_check_password(changeset) do
    with {:ok, user} <-
           UserRepository.find_by_username(changeset.params["username"]),
         {:ok, user} <- confirm_password(user, changeset),
         changeset <- changeset(user, changeset.params) do
      Repo.update changeset
      {:ok, user}
    else
      _ ->
        changeset = changeset
          |> change
          |> add_error(:username, "invalid credentials")
          |> add_error(:password, "invalid credentials")
        {:error, changeset}
    end
  end

  @spec confirm_password(User, Ecto.Changeset) :: {}
  def confirm_password(user, changeset) do
    case Bcrypt.checkpw(changeset.params["password"], user.hashed_password) do
      true ->
        {:ok, user}

      false ->
        {:error, changeset}
    end
  end
end
