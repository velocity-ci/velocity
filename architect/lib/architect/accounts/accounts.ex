defmodule Architect.Accounts do
  @moduledoc """
  The Users context.
  """

  use Guardian, otp_app: :architect

  import Ecto.Query, warn: false
  alias Architect.Repo

  alias Architect.Accounts.User
  alias Comeonin.Bcrypt

  @doc """
  Gets unique id of token subject
  """
  def subject_for_token(user, _claims) do
    {:ok, to_string(user.id)}
  end

  @doc """
  Gets user for token claims
  """
  def resource_from_claims(%{"sub" => id}) do
    case get_user!(id) do
      nil -> {:error, :resource_not_found}
      user -> {:ok, user}
    end
  end

  @doc """
  Authenticate user with username and password
  """
  def authenticate(username, password) when is_binary(username) and is_binary(password) do
    Repo.get_by(User, username: username)
    |> authenticate(password)
  end

  def authenticate(%User{password: actual} = user, password) when is_binary(password) do
    if Bcrypt.checkpw(password, actual) do
      {:ok, user}
    else
      {:error, :invalid_credentials}
    end
  end

  def authenticate(_, _) do
    Bcrypt.dummy_checkpw()
    {:error, :invalid_credentials}
  end

  @doc """
  Returns the list of users.

  ## Examples

      iex> list_users()
      [%User{}, ...]

  """

  def list_users do
    Repo.all(User)
  end

  @doc """
  Gets a single user.

  Raises `Ecto.NoResultsError` if the User does not exist.

  ## Examples

      iex> get_user!(123)
      %User{}

      iex> get_user!(456)
      ** (Ecto.NoResultsError)

  """
  def get_user!(id), do: Repo.get!(User, id)

  @doc """
  Creates a user.

  ## Examples

      iex> create_user(%{field: value})
      {:ok, %User{}}

      iex> create_user(%{field: bad_value})
      {:error, %Ecto.Changeset{}}

  """
  def create_user(attrs \\ %{}) do
    %User{}
    |> User.changeset(attrs)
    |> Repo.insert()
  end
end
