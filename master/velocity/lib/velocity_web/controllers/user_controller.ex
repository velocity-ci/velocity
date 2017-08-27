defmodule VelocityWeb.UserController do
  @moduledoc "provides user management functions.\n"

  use VelocityWeb, :controller
  use Guardian.Phoenix.Controller
  alias Velocity.User
  alias Velocity.ChangesetView
  alias Velocity.UserValidator

  def index(conn, _params) do
    render conn, "index.html"
  end

  @doc "Handles creating a user via POST /users\n"
  @spec create(Conn, {}, {}, {}) :: nil
  def create(conn, user_params, _user, _claims) do
    with changeset <- User.changeset(%User{}, user_params),
         {:ok, changeset} <- UserValidator.register(changeset),
         {:ok, user} <- User.register(changeset) do
      # |> put_resp_header("location", user_path(conn, :show, user))
      conn
      |> put_status(:created)
      |> render("show.json", user: user)
    else
      {:error, changeset} ->
        conn
        |> put_status(:unprocessable_entity)
        |> render(ChangesetView, "error.json", changeset: changeset)
    end
  end
end
