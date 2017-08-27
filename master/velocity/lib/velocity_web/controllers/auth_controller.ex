defmodule VelocityWeb.AuthController do
  @moduledoc "provides authentication functions.\n"

  use VelocityWeb, :controller
  alias Velocity.User
  alias Guardian.Plug
  alias Velocity.ChangesetView
  alias Velocity.UserValidator

  def index(conn, _params) do
    render conn, "index.html"
  end

  @doc "Handles log in via POST /auth\n"
  @spec create(Conn, {}) :: nil
  def create(conn, user_params) do
    with changeset <- User.changeset(%User{}, user_params),
         {:ok, changeset} <- UserValidator.login(changeset),
         {:ok, user} <- User.find_and_check_password(changeset) do
      new_conn = Plug.api_sign_in(conn, user)
      jwt = Plug.current_token(new_conn)
      {:ok, claims} = Plug.claims(new_conn)
      exp = Map.get(claims, "exp")
      conn
      |> put_status(:ok)
      |> render("auth.json", user: user, jwt: jwt, exp: exp)
    else
      {:error, changeset} ->
        conn
        |> put_status(:unauthorized)
        |> render(ChangesetView, "error.json", changeset: changeset)
    end
  end
end
