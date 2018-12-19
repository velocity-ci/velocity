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
  @spec authenticate(String.t(), String.t()) :: {:ok, User.t()} | {:error, String.t()}
  def authenticate(username, password) when is_binary(username) and is_binary(password) do
    user = Repo.one(User, username: username)

    if check_password(user, password) do
      {:ok, user}
    else
      {:error, :invalid_credentials}
    end
  end

  def authenticate(_, _), do: {:error, :invalid_credential_types}

  @spec check_password(User.t(), String.t()) :: boolean()
  defp check_password(nil, _password) do
    Bcrypt.dummy_checkpw()
    false
  end

  defp check_password(user, password), do: Bcrypt.checkpw(password, user.password)

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

  def get_by_username(username) do
    case Repo.one(User, username: username) do
      nil ->
        {:error, :not_found}

      user ->
        {:ok, user}
    end
  end

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

  @doc """
  Updates a user.

  ## Examples

      iex> update_user(user, %{field: new_value})
      {:ok, %User{}}

      iex> update_user(user, %{field: bad_value})
      {:error, %Ecto.Changeset{}}

  """
  def update_user(%User{} = user, attrs) do
    user
    |> User.changeset(attrs)
    |> Repo.update()
  end

  @doc """
  Deletes a User.

  ## Examples

      iex> delete_user(user)
      {:ok, %User{}}

      iex> delete_user(user)
      {:error, %Ecto.Changeset{}}

  """
  def delete_user(%User{} = user) do
    Repo.delete(user)
  end

  @doc """
  Returns an `%Ecto.Changeset{}` for tracking user changes.

  ## Examples

      iex> change_user(user)
      %Ecto.Changeset{source: %User{}}

  """
  def change_user(%User{} = user) do
    User.changeset(user, %{})
  end

  def ensure_admin() do
    case get_by_username("admin") do
    end

    if get_by_username("admin") == {:error} do
      create_user(%{
        username: "admin",
        password: "admin"
      })
    end
  end
end
