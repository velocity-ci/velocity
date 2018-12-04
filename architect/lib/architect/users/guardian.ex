defmodule Architect.Users.Guardian do
  use Guardian, otp_app: :architect

  alias Architect.Users
  alias Comeonin.Bcrypt

  def subject_for_token(user, _claims) do
    {:ok, to_string(user.id)}
  end

  def resource_from_claims(%{"sub" => id}) do
    case Users.get_user!(id) do
      nil -> {:error, :resource_not_found}
      user -> {:ok, user}
    end
  end

  def authenticate_user(changeset) do
    case Users.get_by_username(changeset.params["username"]) do
      {:error} ->
        Bcrypt.dummy_checkpw()
        {:error, :invalid_credentials}

      {:ok, user} ->
        if Bcrypt.checkpw(changeset.params["password"], user.password) do
          {:ok, user}
        else
          {:error, :invalid_credentials}
        end
    end
  end
end
